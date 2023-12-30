import { sleep, check, group } from 'k6';
import { SharedArray } from 'k6/data';
import { Agent } from 'k6/x/immune'
import http from 'k6/http';

export let options = {
  stages: [
    { duration: '1s', target: 1 },
    { duration: '1s', target: 0 },
  ],
};

const fixtures = new SharedArray('fixtures', () => [
  '../apisrv/test/sr630.evidence.json',
  '../apisrv/test/IMN-DELL.evidence.json',
  '../apisrv/test/IMN-SUPERMICRO.evidence.json',
  '../apisrv/test/ludmilla.evidence.json',
  '../apisrv/test/test.evidence.json',
].map((p) => JSON.parse(open(p))));
const server = 'https://xxx.xxx.xxx/v2'
const ca = `
-----BEGIN CERTIFICATE-----
MIIDBjCCAe6gAwIBAgIJALQXxZRsxHr0MA0GCSqGSIb3DQEBCwUAMBgxFjAUBgNV
BAMMDSJrOSBhZ2VudCBDQSIwHhcNMjMwMjE1MTAzMTIzWhcNMzMwMjEyMTAzMTIz
WjAYMRYwFAYDVQQDDA0iazkgYWdlbnQgQ0EiMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAtDjSKTPTBivvz041a2Tvy/dbYo3Fa4W/wgkzH5uVI15hdx4u
9xLTRtDxKPSei/br0RvxTTFrjj+hTGMrYZ5kLNCWmz6TlTNOzptRUQkfP2TIKMk2
ktPGsb8C+oRod/6QVZ2S0mVfoucUgHQZso+4cK0bqVejTFOdiXYoWL0aXJDhZHrO
Gi8n/YA4J02hegwmN5D6RV4thlYQHwSbzXCy/QYaQyD0yKtXtHOo82zj4OqHnyAe
ry+rYmBJYqxv3O8UnmX/+UOAlIYOMhCYkay/zTgYn4TImQQFrMNU6wSUWlH9AK3t
M7aiJ/T1gjRS0HGSZ4CnzTw0PXL3rypVY+Hj2wIDAQABo1MwUTAdBgNVHQ4EFgQU
mqVwYSR9IddY+uFKk8+U7RmZPjEwHwYDVR0jBBgwFoAUmqVwYSR9IddY+uFKk8+U
7RmZPjEwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAmXcmgGZ4
7+EGUhTBJ6LgRdr8qau5F2BKusQv7rpZx3K6zgrJsT4zCmv3h/As5aFADAebnSpE
ecWflfLtDlBGN122rBy1rJg9S5yqJmUdBOqGorCoeKMhXVFGbbJJF+TUtRSQBo3o
0CEnFN0cZtJtYDNJpQq3/vJdlTPlaRvmKnlHfC997zU80u/eZoUNzdq/dIqcsXQa
oNuypgMpfDFixu3o8sofKKMQv/JhTmiB6lZMHsClS3K/J8Uecgfyh3ErJjqQKOQj
spONfm9kM6OWj7vxgoXlhdqDxP/kzvOFjFtsLst0iPe6UDRpBtaa9rCsweQqrCI3
u5DCqK+EDSdlVw==
-----END CERTIFICATE-----
`
// op read -o ca.key 'op://Devops/k9 CA certificate and key/notesPlain'
const key = open( './ca.key');
const enrollmentToken = ''

export default function () {
  const agent = new Agent({
    server: server,
    ca: ca,
    key: key,
  });

  const config = http.get(`${server}/configuration`);
  check(config, {
    'is status 200': (r) => r.status === 200,
    'has no error': (r) => r.json()['errors'] === undefined,
  })

  let token;
  group('enroll', () => {
    const enrollment = agent.createKeys({
      nameHint: 'aa',
      configuration: config.body,
    })
    const credentials = http.post(`${server}/enroll`, enrollment, {
      headers: { 'authorization': `Bearer ${enrollmentToken}` }
    });
    check(credentials, {
      'is status 200': (r) => r.status === 200,
      'has no error': (r) => r.json()['errors'] === undefined,
    })

    token = agent.activateCredentials({
      credentials: credentials.body,
    })
    check(token, { 'has token': (r) => r !== '', })
  })
 
  const fixtureIdx = Math.floor(Math.random() * (fixtures.length - 1));
  const firmware = fixtures[fixtureIdx];
  for (let i = 0; i < 1; i += 1) {
    group('attest', () => {
      const quote = agent.quote({
        firmware: JSON.stringify(firmware),
      })

      const appraisal = http.post(`${server}/attest`, quote, {
        headers: { authorization: `Bearer ${token}` },
      })
      check(appraisal, {
        'is status 200': (r) => r.status === 200,
        'has no error': (r) => r.json()['errors'] === undefined,
        'is trusted': (r) => {
          const d = r.json()
          let data = d['data']
          if (!data) return false
          let attrs = data['attributes']
          if (!attrs) return false
          let verdict = data['verdict']
          if (!verdict) return false
          return verdict['result'] === 'trusted'
        },
      })
    })

    sleep(Math.random() * 10);
  }
}
