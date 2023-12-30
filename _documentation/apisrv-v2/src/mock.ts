import bsearch from 'binary-search-bounds'
import { Request, Response } from 'express'

import { devices, policies } from './database'

const DefaultActor = "42"

export type DeviceState = "new" | "unseen" | "vulnerable" | "trusted" | "outdated" | "retired" | "resurrectable"
export type Device = {
  // writable
  name: string
  attributes: Record<string, string>

  // can only PATCH'd to "retired"
  state: DeviceState

  // writeable once
  cookie: string

  // read only
  id: string
  hwid: string

  replaces: string[]
  replaced_by: string[]
  policies: string[]

  changes: Change[]
  appraisals: Appraisal[]
}

export type Policy = {
  // writable
  name: string
  devices: string[]

  valid_from?: string
  valid_until?: string

  // writable once
  pcr_template?: string[]
  fw_template?: string[]
  cookie: string

  pcrs?: Record<number, string> // PCR -> sha256
  fw_overrides?: string[] // Annotation::id

  // read only
  id: string
  revoked: boolean
  changes: Change[]
}

export type ChangeType = 'enroll' | 'resurrect' | 'rename' | 'tag' | 'associate' | 'template' | 'new' | 'instanciate' | 'revoke' | 'retire'
type Change = {
  type: ChangeType,
  timestamp: string,
  actor: string,
  comment?: string,
}

type Appraisal = {
  verdict: boolean
  report: Record<string, unknown>
}

type DbItem = {
  id: string
}
function dbCmp(a: DbItem, b: DbItem): number {
  return parseInt(b.id, 10) - parseInt(a.id, 10)
}

devices.sort(dbCmp)
policies.sort(dbCmp)

function fetchPolicy(id: string): [Policy, number] | null {
  const idx = bsearch.eq<DbItem>(policies, {id}, dbCmp) as number
  return [policies.length > idx ? policies[idx] : null, idx]
}

function fetchDevice(id: string): [Device, number] | null {
  const idx = bsearch.eq<DbItem>(devices, {id}, dbCmp) as number
  return [devices.length > idx ? devices[idx] : null, idx]
}

function attest(_dev: Device, report: Record<string, unknown>): Appraisal {
  return { report, verdict: true }
}

// helper
function notFound(res: Response, kind: string, id: string) {
  res.statusCode = 404
  res.end(JSON.stringify({
    code: 'err',
    errors: [{
      id: 'oob',
      msg: `no such ${kind} ${id}`,
    }]
  }))
}

function invalidIter(res: Response, iter: unknown) {
  res.statusCode = 400
  res.end(JSON.stringify({
    code: 'err',
    errors: [{
      id: 'inv',
      msg: `invalid iterator ${iter}`,
    }]
  }))
}

function invalidBody(res: Response, path?: string) {
  res.statusCode = 400
  res.end(JSON.stringify({
    code: 'err',
    errors: [{
      id: 'inv',
      path,
      msg: `invalid request`,
    }]
  }))
}

function logicError(res: Response, msg: string, path?: string) {
  res.statusCode = 400
  res.end(JSON.stringify({
    code: 'err',
    errors: [{
      id: 'log',
      path,
      msg,
    }]
  }))
}

function newId(ary: DbItem[]): string {
  while (true) {
    const id = Math.trunc(Math.random() * 100000).toString()
    if (bsearch.eq(ary, {id}) === -1) {
      return id
    }
  }
}


// GET /v2/devices?i=...
export function getDevices(req: Request, res: Response) {
  res.setHeader('Content-Type', 'application/json')

  const iter = parseInt(req.query.i as string || '0', 10)
  const dev = devices.length > iter ? devices[iter] : null
  if (!dev) {
    return invalidIter(res, iter)
  }

  res.end(JSON.stringify({
    code: 'ok',
    data: {
      devices: [dev],
    },
    meta: devices.length > iter + 1 ? { next: (iter + 1).toString() } : undefined
  }))
}

// GET /v2/policies?i=...
export function getPolicies(req: Request, res: Response) {
  res.setHeader('Content-Type', 'application/json')

  const iter = parseInt(req.query.i as string || '0', 10)
  const pol = policies.length > iter ? policies[iter] : null
  if (!pol) {
    return invalidIter(res, iter)
  }

  res.end(JSON.stringify({
    code: 'ok',
    data: {
      policies: [pol],
    },
    meta: policies.length > iter + 1 ? { next: (iter + 1).toString() } : undefined
  }))
}

// GET /v2/devices/:id
export function getDevice(req: Request, res: Response) {
  res.setHeader('Content-Type', 'application/json')

  const [dev, _] = fetchDevice(req.params.id)
  if (!dev) {
    return notFound(res, "device", req.params.id)
  }

  res.end(JSON.stringify({
    code: 'ok',
    data: {
      devices: [dev],
    }
  }))
}

// GET /v2/policies/:id
export function getPolicy(req: Request, res: Response) {
  res.setHeader('Content-Type', 'application/json')

  const [pol, _] = fetchPolicy(req.params.id)
  if (!pol) {
    return notFound(res, "policy", req.params.id)
  }

  res.end(JSON.stringify({
    code: 'ok',
    data: {
      policies: [pol],
    }
  }))
}

// GET /v2/info
export function getInfo(_: Request, res: Response) {
  res.end(JSON.stringify({
    code: 'ok',
    data: {
      info: {
        api_version: "2"
      }
    }
  }))
}

// PATCH /v2/devices/:id
export function patchDevice(req: Request, res: Response) {
  res.setHeader('Content-Type', 'application/json')

  const [dev, idx] = fetchDevice(req.params.id)
  if (!dev) {
    return notFound(res, "device", req.params.id)
  }

  let patchRaw = ''
  req.on('data', chunk => {
    patchRaw += chunk;
  })
  req.on('end', () => {
    let patch
    try {
      patch = JSON.parse(patchRaw)
    } catch {
      return invalidBody(res, "/")
    }

    if (!(patch instanceof Object)) {
      return invalidBody(res, "/")
    }

    // read patch
    for (const [field, value] of Object.entries(patch)) {
      switch (field) {
        case "name":
          if (typeof value !== "string" || value === "") {
            return invalidBody(res, "/name")
          }
          dev.name = patch.name
          break;

        case "attributes":
          if (!(value instanceof Object))  {
            return invalidBody(res, "/attributes")
          }

          for (const [attr, av] of Object.entries(value)) {
            if (av === null) {
              delete dev.attributes[attr]
            } else if (typeof av !== "string") {
              return invalidBody(res, `/attributes/${attr}`)
            } else {
              dev.attributes[attr] = av as string
            }
          }
          break;

        case "state":
          if (value !== dev.state && value !== "retired")  {
            return invalidBody(res, "/state")
          }
          dev.state = value as DeviceState
          break

        default:
          return invalidBody(res, `/${field}`)
      }
    }

    devices[idx] = dev

    res.end(JSON.stringify({
      code: 'ok',
      data: {
        devices: [dev],
      }
    }))
  })
}

// POST /v2/devices/:id/resurect => Device
export function resurectDevice(req: Request, res: Response) {
  res.setHeader('Content-Type', 'application/json')

  const [dev, devidx] = fetchDevice(req.params.id)
  if (!dev) {
    return notFound(res, "device", req.params.id)
  }

  if (dev.state !== "retired") {
    return logicError(res, "device is not retired")
  }

  if (dev.replaced_by && dev.replaced_by.length > 0) {
    return logicError(res, "device cannot be resurected")
  }

  const newDev = Object.assign({}, dev) as Device
  newDev.id = newId(devices)
  newDev.name = `Device #${newDev.id}`
  newDev.replaces = [dev.id]
  newDev.replaced_by = []
  newDev.policies = []
  newDev.appraisals = []
  newDev.changes = [{
    timestamp: Date.now().toString(),
    actor: DefaultActor,
    type: "resurrect",
    comment: undefined,
  }]

  dev.replaced_by = [newDev.id]
  dev.changes.push({
    timestamp: Date.now().toString(),
    actor: DefaultActor,
    type: "retire",
    comment: undefined,
  })


  const changedPolices = []
  const changedDevices = []

  // update policy cross refs
  for (const id of dev.policies) {
    const [pol, idx] = fetchPolicy(id)
    if (pol) {
      pol.devices.push(newDev.id)
      policies[idx] = pol
      newDev.policies.push(pol.id)
      changedPolices.push(pol)
    }
  }

  // reattest
  if (dev.appraisals.length > 0) {
    const appr = attest(newDev, dev.appraisals[dev.appraisals.length - 1].report)
    newDev.appraisals.push(appr)

    if (appr.verdict) {
      newDev.state = "trusted"
    } else {
      newDev.state = "vulnerable"
    }
  } else {
    newDev.state = "unseen"
  }

  devices[devidx] = dev
  devices.push(newDev)
  devices.sort(dbCmp)
  changedDevices.unshift(newDev)
  changedDevices.push(dev)

  res.end(JSON.stringify({
    code: 'ok',
    data: {
      devices: changedDevices,
      policies: changedPolices
    }
  }))
}

// POST /v2/policies
export function postPolicy(req: Request, res: Response) {
  res.setHeader('Content-Type', 'application/json')

  let policyRaw = ''
  req.on('data', chunk => {
    policyRaw += chunk;
  })
  req.on('end', () => {
    let pol
    try {
      pol = JSON.parse(policyRaw)
    } catch {
      return invalidBody(res, "/")
    }

    if (!(pol instanceof Object)) {
      return invalidBody(res, "/")
    }

    for (const [field, value] of Object.entries(pol)) {
      switch (field) {
          // non empty string
        case "name":
        case "cookie":
          if (typeof value !== "string" || value === "") {
            return invalidBody(res, `/${field}`)
          }
          break

          // non empty numeric string
        case "valid_from":
        case "valid_until":
        if (typeof value !== "string" || !/\d+/g.test(value)) {
            return invalidBody(res, `/${field}`)
          }
          break

          // string[] to existing devices
        case "devices":
          if (!(value instanceof Array)) {
            return invalidBody(res, "/devices")
          }
          for (const [idx, val] of value.entries()) {
            if (typeof val !== "string" || !fetchDevice(val)) {
              return invalidBody(res, `/devices[${idx}]`)
            }
          }
          break

          // string[] with decimal numbers 0-26
        case "pcr_template":
          if (!(value instanceof Array)) {
            return invalidBody(res, "/pcr_template")
          }
          for (const [idx, val] of value.entries()) {
            if (typeof val !== "string" || !/\d+/g.test(val)) {
              return invalidBody(res, `/pcr_template[${idx}]`)
            }
          }
          break

          // Map<string, string> with decimal numbers 0-26 to hex values
        case "pcrs":
          if (!(value instanceof Map)) {
            return invalidBody(res, "/pcrs")
          }
          for (const [key, pcr] of value) {
            if (typeof key !== "string" || !/\d+/g.test(key)) {
              return invalidBody(res, `/pcrs[${key}]`)
            }
            if (typeof pcr !== "string" || !/([a-fA-F0-9]{2})+/g.test(pcr)) {
              return invalidBody(res, `/pcrs[${key}]`)
            }
          }
          break

          // string[]
        case "fw_overrides":
          if (!(value instanceof Array)) {
            return invalidBody(res, "/fw_overrides")
          }
          for (const [idx, val] of value.entries()) {
            if (typeof val !== "string") {
              return invalidBody(res, `/fw_overrides[${idx}]`)
            }
          }
          break

        default:
      }
    }

    // cross field sanity checks
    if (pol.valid_until && pol.valid_since) {
      if (parseInt(pol.valid_until, 10) < parseInt(pol.valid_since, 10)) {
        return invalidBody(res, "/valid_since")
      }
    }
    if (pol.pcr_template && pol.pcrs) {
      return invalidBody(res, "/pcr_template")
    }

    const devs = []

    // check whether we've done this before
    if (pol.cookie) {
      for (const pol2 of policies) {
        if (pol2.cookie === pol.cookie) {
          for (const dev of pol2.devices) {
            const d = fetchDevice(dev)
            if (d) {
              devs.push(d)
            }
          }

          res.statusCode = 202
          res.end(JSON.stringify({
            code: 'ok',
            data: {
              devices: devs,
              policies: [pol2]
            }
          }))
          return
        }
      }
    }

    pol.id = newId(policies)
    policies.push(pol)
    policies.sort(dbCmp)

    // update cross refs & reattest
    for (const dev of pol.devices) {
      const [d, idx] = fetchDevice(dev)
      if (d) {
        // cross ref
        d.policies.push(pol.id)

        // reattest
        if (pol.pcrs && d.state !== "retired" && d.appraisals.length > 0) {
          const report = d.appraisals[d.appraisals.length - 1].report
          const appr = attest(d, report)

          d.appraisals.push(appr)
          if (appr.verdict) {
            d.state = "trusted"
          } else {
            d.state = "vulnerable"
          }
        }

        devices[idx] = d
        devs.push(d)
      }
    }

    res.end(JSON.stringify({
      code: 'ok',
      data: {
        devices: devs,
        policies: [pol]
      }
    }))
  })
}

// DELETE /v2/policies/:id
export function deletePolicy(req: Request, res: Response) {
  res.setHeader('Content-Type', 'application/json')

  const [pol, idx] = fetchPolicy(req.params.id)
  if (!pol) {
    res.statusCode = 202
    res.end(JSON.stringify({
      code: 'ok',
      data: {
        policies: [pol],
      }
    }))
  } else {
    policies.splice(idx, 1)

    res.end(JSON.stringify({
      code: 'ok',
      data: {
        policies: [pol],
      }
    }))
  }
}
