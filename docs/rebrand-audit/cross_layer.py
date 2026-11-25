#!/usr/bin/env python3
"""Cross-layer reference integrity for the rebrand."""
import re, json
from pathlib import Path
from collections import defaultdict
ROOT = Path('/data/workspace/mit-assets-admin')
ok = lambda m: print(f"  ✓ {m}")
fail = lambda m: print(f"  ✗ {m}")

def read(p): return (ROOT/p).read_text(encoding='utf-8', errors='replace')

print("=== A. Go: every global.MAA_* reference resolves to a declaration in package global ===")
gs = read('server/global/global.go') + read('server/global/model.go')
declared = set(re.findall(r'^\s*(?:var\s+)?(MAA_[A-Za-z0-9_]+)\s', gs, re.M)) | set(re.findall(r'^type (MAA_[A-Za-z0-9_]+) struct', gs, re.M))
used = set()
for p in (ROOT/'server').rglob('*.go'):
    if '/docs/' in p.as_posix(): continue
    used |= set(re.findall(r'global\.(MAA_[A-Za-z0-9_]+)', p.read_text(encoding='utf-8', errors='replace')))
missing = used - declared
(ok if not missing else fail)(f"declared={len(declared)} used={len(used)} missing={sorted(missing) or 'none'}")

print("\n=== B. Go: no residual GVA_/Gva/GVA identifiers anywhere (excluding DB table names) ===")
res = []
for p in (ROOT/'server').rglob('*.go'):
    t = p.read_text(encoding='utf-8', errors='replace')
    for i, line in enumerate(t.splitlines(), 1):
        if 'gva_announcement' in line: continue
        if re.search(r'\bGVA_|\bGva[A-Z]|\bGVA[A-Z]|\bgva_', line):
            res.append(f"{p.relative_to(ROOT)}:{i}: {line.strip()[:100]}")
(ok if not res else fail)(f"residual Go identifiers: {len(res)}")
for r in res[:10]: print("     ", r)

print("\n=== C. env var contract: compose passes ⊆ .env / .env.example keys, Go reads ⊆ compose ∪ .env ===")
env_keys = set(re.findall(r'^([A-Z_]+)=', read('deploy/docker-dev/.env'), re.M))
example_keys = set(re.findall(r'^([A-Z_]+)=', read('deploy/docker-dev/.env.example'), re.M))
compose_refs = set(re.findall(r'\$\{?([A-Z_]+)', read('deploy/docker-dev/docker-compose.yml')))
go_reads = set()
for p in (ROOT/'server').rglob('*.go'):
    if '/docs/' in p.as_posix(): continue
    go_reads |= set(re.findall(r'os\.Getenv\("([A-Z_]+)"\)', p.read_text(encoding='utf-8', errors='replace')))
sh_reads = set()
for p in (ROOT/'deploy/docker-dev').rglob('*.sh'):
    sh_reads |= set(re.findall(r'\$\{?([A-Z_]+)', p.read_text(encoding='utf-8', errors='replace')))
print(f"  .env={len(env_keys)} .env.example={len(example_keys)} compose-refs={len(compose_refs)} go-reads={len(go_reads)} sh-reads={len(sh_reads)}")
# .env 与 example 键集合差异
d1 = env_keys - example_keys; d2 = example_keys - env_keys
(ok if not d1 and not d2 else print)(f"  .env vs example diff: only-in-env={sorted(d1)} only-in-example={sorted(d2)}")
# go 读取的 MAA_ 变量是否都来自 compose 或 .env
go_maa = {k for k in go_reads if k.startswith('MAA_')} - {'MAA_MCP_AUTH_TOKEN','MAA_MCP_BIN','MAA_MCP_BIND','MAA_MCP_CONFIG'}  # internal parent->child env
unreachable = go_maa - compose_refs - env_keys - example_keys
(ok if not unreachable else fail)(f"Go getenv MAA_* not provided anywhere: {sorted(unreachable) or 'none'}")
# 任何地方还引用 GVA_ 变量
stale = {k for k in (compose_refs | sh_reads | go_reads) if k.startswith('GVA_')}
(ok if not stale else fail)(f"stale GVA_ var references: {sorted(stale) or 'none'}")

print("\n=== D. JWT: audience/issuer consistent between issuer and verifier ===")
jwt = read('server/utils/jwt.go')
aud = re.findall(r'ClaimStrings\{"([A-Z]+)"\}', jwt)
iss = re.findall(r'Issuer:\s+global\.MAA_CONFIG\.JWT\.Issuer', jwt)
(ok if aud == ['MAA'] and iss else fail)(f"audience={aud} issuer-from-config={bool(iss)}")
# API token 签发处也要用同一 audience
tok = read('server/service/system/sys_api_token.go')
aud2 = re.findall(r'ClaimStrings\{"([A-Z]+)"\}', tok)
(ok if aud2 == ['MAA'] else fail)(f"api-token audience={aud2}")

print("\n=== E. MCP tool names: registered == referenced in prompts == frontend docs ===")
reg = set(re.findall(r'mcp\.NewTool\("([a-z_]+)"', read('server/mcp/maa_execute.go')+read('server/mcp/maa_review.go')+read('server/mcp/maa_analyze.go')))
allmcp = ''.join(p.read_text(encoding='utf-8') for p in (ROOT/'server/mcp').rglob('*.go'))
refs = set(re.findall(r'\b(maa_(?:analyze|execute|review|auto_generate))\b', allmcp))
stale_refs = set(re.findall(r'\b(gva_(?:analyze|execute|review|auto_generate))\b', allmcp))
(ok if not stale_refs else fail)(f"registered={sorted(reg)} referenced={sorted(refs)} stale={sorted(stale_refs) or 'none'}")
(ok if refs <= reg | {'maa_auto_generate'} else fail)(f"all referenced tools registered: {refs <= reg | {'maa_auto_generate'}}")

print("\n=== F. Frontend: maaModel json field ↔ backend tag ===")
be = re.findall(r'json:"(maaModel)"', read('server/model/system/request/sys_auto_code.go'))
fe = len(re.findall(r'\bmaaModel\b', read('web/src/view/systemTools/autoCode/index.vue')))
(ok if be and fe else fail)(f"backend tag={be} frontend uses={fe}")

print("\n=== G. Frontend: iconfont family name ↔ class ↔ usages ===")
css = read('web/src/style/iconfont.css')
fam = set(re.findall(r"font-family: '([A-Za-z]+)'", css))
cls = set(re.findall(r'^\.([A-Za-z]+)\s*\{', css, re.M))
used_cls = set()
for p in (ROOT/'web/src').rglob('*'):
    if p.suffix in ('.vue', '.js'):
        used_cls |= set(re.findall(r'\b(maaIcon[A-Za-z-]*|gvaIcon[A-Za-z-]*)\b', p.read_text(encoding='utf-8', errors='replace')))
stale = {c for c in used_cls if c.startswith('gva')}
(ok if fam == {'maaIcon'} and not stale else fail)(f"family={fam} class={cls} used={len(used_cls)} stale-gva={sorted(stale) or 'none'}")

print("\n=== H. Frontend: route name MaaLayoutIframe consistent ===")
names = set()
for p in (ROOT/'web/src').rglob('*'):
    if p.suffix in ('.vue', '.js', '.json'):
        names |= set(re.findall(r'\b([A-Za-z]*LayoutIframe)\b', p.read_text(encoding='utf-8', errors='replace')))
(ok if names == {'MaaLayoutIframe'} else fail)(f"variants found: {sorted(names)}")

print("\n=== I. CSS class mapping completeness: every .maa-* used in templates has a style or is a semantic hook ===")
used = set()
for p in (ROOT/'web/src').rglob('*'):
    if p.suffix in ('.vue', '.js', '.html'):
        used |= set(re.findall(r'\b(maa-[a-z0-9-]+)\b', p.read_text(encoding='utf-8', errors='replace')))
defined = set()
for p in (ROOT/'web/src').rglob('*'):
    if p.suffix in ('.vue', '.scss', '.css'):
        defined |= set(re.findall(r'\.(maa-[a-z0-9-]+)', p.read_text(encoding='utf-8', errors='replace')))
undefined = used - defined
print(f"  used={len(used)} defined={len(defined)} used-but-undefined={sorted(undefined)}")
print("  (undefined 的是语义钩子或由 Element/UnoCSS 处理，只要与备份中 gva-* 集合一一对应即可)")
# 对照备份
bk = ROOT/'.deploy/backups/20260910131409-86ed705/web/src'
bused = set()
for p in bk.rglob('*'):
    if p.suffix in ('.vue', '.js', '.html'):
        bused |= set(re.findall(r'\b(gva-[a-z0-9-]+)\b', p.read_text(encoding='utf-8', errors='replace')))
bdef = set()
for p in bk.rglob('*'):
    if p.suffix in ('.vue', '.scss', '.css'):
        bdef |= set(re.findall(r'\.(gva-[a-z0-9-]+)', p.read_text(encoding='utf-8', errors='replace')))
bundef = {c[4:] for c in bused - bdef}
(ok if {c[4:] for c in undefined} == bundef else fail)(f"undefined-set matches baseline: {({c[4:] for c in undefined} == bundef)}")
