# Repository operating instructions

## Repository location

- The real Git repository is `D:\资产管理系统\gin-vue-admin`.
- `D:\gin-vue-admin` is only a workspace shell and must not be treated as the source repository.

## Mandatory verification after code changes

- Preserve unrelated user changes and stage only files that belong to the current task.
- Run the narrowest relevant tests first, then an appropriate build or compile check.
- `git diff --check` must pass before publishing.
- Existing unrelated failures may be reported and bypassed only when task-specific checks pass and the failure is demonstrably outside the changed scope.

## Automatic publish and production restart

The user has explicitly authorized automatic publication and production restart after completed code changes. Once an implementation is finished and verified, do the following without asking for a second confirmation:

1. Confirm the repository is on `main` and inspect `git status -sb` plus the complete task diff.
2. Commit only the task files with a concise conventional commit message.
3. Push the commit to `origin/main` using Git SSH.
4. Deploy that exact committed revision to production and rebuild every affected service.
5. Run the production health check and inspect container status before reporting success.

Do not auto-publish when the user requested only diagnosis, review, explanation, or a local experiment. Do not publish from a non-`main` branch; report the branch mismatch instead.

## Production deployment facts

- SSH is available from WSL through the alias `gin-vue-admin-remote`.
- Production root: `/data/gin-vue-admin`.
- The production root is a source deployment tree without `.git`; `git pull` will fail there.
- Deployment staging: `/data/gin-vue-admin/.deploy/staging`.
- Deployment backups: `/data/gin-vue-admin/.deploy/backups`.
- Revision marker: `/data/gin-vue-admin/.deploy/current-commit`.
- Compose directory: `/data/gin-vue-admin/deploy/docker-dev`.

Publish with an archive created from the committed Git revision, never from an uncommitted working tree. Upload it through the WSL SSH alias, create a timestamped backup of the affected source directories, extract the archive over `/data/gin-vue-admin`, and preserve ignored runtime files such as:

- `deploy/docker-dev/.env`
- `deploy/docker-dev/config.yaml`
- everything under `.deploy`

Determine affected services from the files changed since `.deploy/current-commit`. If the marker is absent or unreliable, rebuild both services.

- `server/**` only: run `./build.sh server`, then force-recreate `server`.
- `web/**` only: run `./build.sh web`, then force-recreate `web`.
- Both trees, shared deployment files, or an unknown baseline: build and force-recreate `server web`.

The recreate command is:

```bash
docker compose --env-file .env -f docker-compose.yml up -d --force-recreate <services>
```

`restart.sh` alone is forbidden for code updates because it does not rebuild images. After recreation, run:

```bash
./health-check.sh
./ps.sh
```

Only after both succeed, write the full deployed commit hash to `.deploy/current-commit`. If upload, build, recreation, or health checking fails, keep the previous backup, stop, and report the exact failed stage. Never claim production was updated unless the health check passed.

## Secrets and safety

- Never commit, print, archive, or overwrite `.env`, `config.yaml`, credentials, private keys, database dumps, or object-storage secrets.
- Do not use destructive Git commands or delete deployment backups.
- A deployment must be built from a pushed commit so GitHub and production remain traceable to the same revision.
