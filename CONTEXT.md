# Domain Context

## Asset

An Asset is a physical or accountable item managed through registration, location, custody, valuation, and operational records.

## Invoice

An Invoice is a financial source document uploaded by a user. It owns the original evidence file, recognized fields, monetary totals, classification, review state, and audit metadata.

## Invoice Item

An Invoice Item is a recognized line item belonging to an Invoice, including the item name, specification, quantity, unit price, amount, tax rate, and tax amount.

## Recognition Job

A Recognition Job is a durable request to extract structured Invoice fields from an uploaded evidence file. It records provider choice, attempts, state, timing, and failure details.

## Invoice Category

An Invoice Category is the reporting classification assigned to a confirmed Invoice, such as office procurement, fixed assets, travel, software subscriptions, or maintenance.

## Classification Rule

A Classification Rule is an explainable matching rule that scores recognized supplier, item, invoice type, or raw text against an Invoice Category.

## Review

Review is the human verification step between recognition and confirmation. Only confirmed Invoices are included in official statistics.

## Evidence File

An Evidence File is the private RustFS or MinIO object backing an Invoice. Access always requires authentication and data authorization.

## Evidence Cleanup Job

An Evidence Cleanup Job is a transactional outbox record created when an Invoice is deleted. It retains the original storage location and retries idempotent object deletion under a lease, so database deletion and private-file cleanup remain recoverable across process or storage failures.

## Tenant

A Tenant is the highest business-data ownership boundary. Users, departments, and tenant-owned records never cross this boundary during normal operations.
_Avoid_: Organization, company, customer account

## Department

A Department is a hierarchical organizational unit inside one Tenant. A Department cannot move to another Tenant while retaining its identity.
_Avoid_: Group, team, role

## Department Membership

A Department Membership associates a User with a Department as a primary or additional membership. Membership expresses organizational belonging, not API permission.

## Data Scope

A Data Scope is the business-record visibility granted by a Role, limited to self, department, department subtree, tenant, or platform-wide administration. It is distinct from Casbin API permission.
_Avoid_: Data authority, menu permission

## Platform Administrator

A Platform Administrator is the only actor allowed to operate across Tenant boundaries. Tenant administrators remain restricted to their own Tenant.
