CREATE TYPE orga_suspension_reason AS ENUM (
  'not-suspended',
  'abuse',
  'non-payment',
  'unknown',
)

CREATE TYPE orga_address AS (
  city TEXT NOT NULL,
  country TEXT NOT NULL,
  line1 TEXT NOT NULL,
  line2 TEXT NOT NULL,
  state TEXT NOT NULL,
  zip TEXT NOT NULL,
  name TEXT NOT NULL,
)

CREATE TABLE organizations (
  id BIG SERIAL PRIMARY KEY, /* immutable */
  external UUID NOT NULL,
  name TEXT NOT NULL,
  cookie TEXT NOT NULL, /* immutable */

  plan_id REFERENCES(plan.id) NOT NULL,
  stripe orga_stripe_objects NOT NULL,
  credit int NOT NULL CHECK (credit <= 0),

  suspended: orga_suspension_reason NOT NULL,
)

CHECK ((SELECT COUNT(*) FROM invoice WHERE org_id = ?id AND state = 'preliminary') == 1)

CREATE TABLE plan (
  id BIG SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  default BOOLEAN,

  maximum_slots int NOT NULL CHECK (maximum_slots > 0),
  fixed_price int NULL CHECK (fixed_price >= 0),
  variable_price int NULL CHECK (variable_price >= 0),
)

CREATE VIEW default_organizations AS
  SELECT orga_id, user_id FROM memberships WHERE default = TRUE;
CREATE UNIQUE INDEX default_organizations(orga_id, user_id)

CREATE TYPE plan_update AS (
  day int NOT NULL CHECK (day >= 1 AND day <= 31),
  slots int NOT NULL CHECK (slots >= 0),
)

CREATE TYPE invoice_state AS ENUM (
  'preliminary', /* invoice of the current month */
  'draft', /* invoice finalized */
  'open', /* stripe invoice created, email notifications sent */
  'overdue_level1', /* unpaid for 2 weeks */
  'overdue_level2', /* unpaid for 4 weeks */
  'overdue_level3', /* unpaid for 6 weeks, org suspended */
  'paid',
  'void',
)

CREATE TABLE invoice (
  id BIG SERIAL PRIMARY KEY,
  plan_id REFERENCES(plan.id),
  org_id REFERENCES(organizations.id),

  state invoice_state NOT NULL,
  stripe_invoice TEXT NULL
    CHECK (stripe_invoice IS NULL == (invoice_state IN {'preliminary', 'draft', 'void'})),

  month INTEGER NOT NULL
    CHECK (month <= 1 AND month >= 12), /* 1 (January) - 12 (December) */
  year INTEGER NOT NULL
    CHECK (year >= 2020),

  updates updates ARRAY[] CHECK (array_length(updates) >= 1)
)

CHECK ((SELECT MAX(updates.slots) FROM invoice WHERE plan_id = ?id) <= (SELECT maximum_slots FROM plan WHERE id = ?id))


/*
Invoice

invariants

stripe_invoice == NULL -> invoice_state in {'preliminary', 'draft', 'void'}
len(updates) > 0
updates.slots <= plan_id.maximum_slots
*/


/*
1st of each month
preliminary -> draft
Draft-Invoice(month, year, org_id):
  invoice = SELECT * FROM invoice WHERE
    month = ?month AND
    year = ?year AND
    state = 'draft' AND
    org.id = ?org_id
  invoice.state = 'draft'
  queue Open-Invoice(invoice)

  new_invoice = {
    plan_id = invoice.plan_id,
    org_id = invoice.org_id,

    state = 'draft'
    stripe_invoice = NULL
    month = 1 + ((invoice.month - 1) % 12)
    year = invoice.year + (invoice.month == 12 ? 1 : 0)
    updates = [{
      day = 1,
      slots = invoice.updates[-1].slots
    }]
  }

draft -> open
Open-Invoice(invoice)
  invoice_items = F(invoice.updates, invoice.plan.variable_price, invoice.plan.fixed_price, orga.credit)
  orga.credit -= F(invoice_items)
  invoice.stripe_invoice = Stripe::Create-Invoice(org.stripe_customer, invoice_items)
  queue Email-Invoice(invoice)

cronjob
open -> overdue_level1
Send-1-Warning:
  if invoice.state == 'open':
    invoice.state = 'overdue_level1'
  else:
    fail

cronjob
overdue_level1 -> overdue_level2
Send-2-Warning:
  if invoice.state == 'overdue_level1':
    invoice.state = 'overdue_level2'
  else:
    fail

cronjob
overdue_level1 -> overdue_level3
Send-3-Warning:
  if invoice.state == 'overdue_level2':
    invoice.state = 'overdue_level3'
    orga.suspended = TRUE
    send UpdateQuota
  else:
    fail

Stripe invoice.state == paid
open, overdue_level* -> paid
Pay-Invoice:
  invoice.state = 'paid'
  if invoice.stripe_invoice != NULL:
    Stripe::Pay-Invoice(invoice.stripe_invoice)

Strip invoice.state == void
* -> void
Void-Invoice:
  invoice.state = 'void'
  if invoice.stripe_invoice != NULL:
    Stripe::Void-Invoice(invoice.stripe_invoice)
*/

CREATE TYPE user_auth_provider AS ENUM (
  'google',
  'github',
  'native',
)

CREATE TABLE users (
  id BIG SERIAL PRIMARY KEY,
  cookie TEXT NOT NULL,

  name TEXT NOT NULL,
  email TEXT NOT NULL,

  /* auth */
  totp_secret TEXT NULL,
  password_hash TEXT NULL,
  provider user_auth_provider NULL,
  provider_uid TEXT NULL,

  /* priv. */
  is_admin BOOLEAN NOT NULL,
)

CREATE TYPE membership_role AS ENUM (
  'member',
  'owner'
)

CREATE TABLE memberships (
  user_id REFERENCES(users.id) NOT NULL,
  orga_id REFERENCES(organizations.id) NOT NULL,

  role membership_role NOT NULL,

  send_invoices BOOLEAN NOT NULL,
  send_alerts BOOLEAN NOT NULL,

  default BOOLEAN NOT NULL,
  confirmed BOOLEAN NOT NULL,
)

/*
User

invited: default membership.confirmed == false AND suspended == false
active: default membership.confirmed == true AND suspended == false
suspended: suspended == true

invariants

one default membership
non default membership.confirmed -> default membership.confirmed
name unique
cookie unique
provider_uid == NULL iff provider == 'native'
password_hash != NULL iff provider == 'native'
totp_secret != NULL -> provider == 'native'

Orga

created: stripe_customer == NULL AND suspended == 'not-suspended'
active: stripe_customer != NULL AND suspended == 'not-suspended'
suspended: suspended != 'not-suspended'

invariants

name unique
external unique
cookie unique
billing_addess_dirty == false -> stripe_customer != NULL
*/
