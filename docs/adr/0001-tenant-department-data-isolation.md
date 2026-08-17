---
status: accepted
date: 2026-08-17
---

# Tenant-rooted department data isolation

The system will use Tenant as the highest business-data boundary and a Department tree inside each Tenant for organizational visibility. Existing single-organization data will be assigned to a default Tenant and root Department; Casbin continues to authorize API paths, while a mandatory service-level Data Scope enforces row ownership for every tenant-owned query and mutation. Role relationships will not grant cross-Tenant visibility, and only the Platform Administrator may use the platform-wide scope. This preserves the current internal deployment while avoiding a second disruptive migration if the product later serves multiple organizations.

## Consequences

- Tenant-owned records require `tenant_id`; department-owned records also require `department_id`.
- Client input never decides record ownership. Create operations derive ownership from the authenticated actor.
- List, detail, update, delete, export, file access, AI tools, and background jobs use the same Data Scope.
- Existing `DataAuthorityId` role relationships cannot substitute for Tenant or Department ownership.
- PostgreSQL RLS may be added after application enforcement as defense in depth, but it is not the primary contract because tests and supported development databases must share the same behavior.
