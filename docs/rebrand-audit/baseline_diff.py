#!/usr/bin/env python3
"""
Rebrand audit: take the pre-rename baseline (backup d7e1eea), apply the exact
expected substitutions, and diff against the current tree. Any residual line
must be explainable as post-backup business development or an audit-rule gap.
"""
import os, re, sys, difflib
from pathlib import Path

ROOT = Path('/data/workspace/mit-assets-admin')
BASE = ROOT / '.deploy/backups/20260911145857-d7e1eea'

# exact ordered substitution rules (regex, replacement) applied to baseline text
RULES = [
    # Go identifiers
    (r'\bGVA_', 'MAA_'),
    (r'\bGVA([A-Z])', r'MAA\1'),
    (r'\bGva([A-Z])', r'Maa\1'),
    (r'\bgvaModel\b', 'maaModel'),
    (r'gva默认Model', '项目默认Model'),
    (r'gva-plug-temp', 'maa-plug-temp'),
    (r'gva-mcp-standalone', 'maa-mcp-standalone'),
    (r'gva-mcp"', 'maa-mcp"'),
    (r'gva_auto_generate', 'maa_auto_generate'),
    (r'gva_execute', 'maa_execute'),
    (r'gva_analyze', 'maa_analyze'),
    (r'gva_review', 'maa_review'),
    (r'plugin/gva/', 'plugin/maa/'),
    (r'"gva"', '"maa"'),
    (r'/gva"', '/maa"'),
    (r'gva-assets', 'maa-assets'),
    (r'ClaimStrings\{"GVA"\}', 'ClaimStrings{"MAA"}'),
    (r'Source: "GVA"', 'Source: "MAA"'),
    (r'YOUR_GVA_TOKEN', 'YOUR_MAA_TOKEN'),
    (r'GVAPLUGPINATH', 'MAAPLUGPINATH'),
    # author headers
    (r'//\s*@author:.*', '//@author: [MiTo]'),
    (r'// Author \[[^\]]+\]\([^)]+\)', '// Author [MiTo]'),
    (r'// Author \[[^\]]+\]', '// Author [MiTo]'),
    (r'// Author [A-Za-z0-9_]+\s*$', '// Author MiTo'),
    (r'( \* @Author:).*', r'\1 MiTo'),
    (r'( \* @LastEditors:).*', r'\1 MiTo'),
    (r'^\s*\* @(Date|LastEditTime|FilePath):.*\n', ''),
    # MCP prompts / Chinese comments
    (r'GVA代码', '项目代码'), (r'GVA模型', '项目模型'), (r'GVA框架', '项目框架'),
    (r'GVA生成', '项目代码生成'), (r'GVA分析器', '项目代码分析器'), (r'创建GVA', '创建项目'),
    (r'gva的User', '项目的User'),
    (r'\[MiTo\]\([^)]+\)', '[MiTo]'),
    (r'// 业务库 gva => gva, 不要传"gva"', '// 业务库标识（db-list 的 alias-name），不要带引号'),
    (r'主数据库通常使用 gva（默认数据库标识）', '主数据库使用 config.yaml 中 db-list 配置的 alias-name'),
    # avatars
    (r'https://qmplusimg\.henrongyi\.top/[A-Za-z0-9_]+\.(png|jpg)', '/logo.png'),
    # nav link removal
    (r'^\t\t\{MenuLevel: 0, Hidden: true, ParentId: 0, Path: "https://github\.com/WangMi2022/YuYuan-Pass-System".*\n', ''),
    (r', "https://github\.com/WangMi2022/YuYuan-Pass-System", "plugin"', ', "plugin"'),
    # weak key test
    (r'\{name: "default", value: "qmPlus", want: true\},',
     '{name: "legacy default", value: "qmPlus", want: true},\n\t\t{name: "default", value: "mit-assets-admin", want: true},'),
    (r'case "", "qmplus", "changeme"', 'case "", "qmplus", "mit-assets-admin", "changeme"'),
    (r'jwt\.ClaimStrings\{"GVA"\}', 'jwt.ClaimStrings{"MAA"}'),
    # frontend
    (r'\bgva-', 'maa-'),
    (r'\bgvaIcon', 'maaIcon'),
    (r"'gvaIcon'", "'maaIcon'"),
    (r'\bgvaLayoutIframe', 'MaaLayoutIframe'),
    (r'GvaLayoutIframe', 'MaaLayoutIframe'),
    (r'useGva\b', 'useMaa'),
    (r'gvaAside', 'maaAside'),
    (r'\bGva([A-Z])', r'Maa\1'),
    (r'gva_ai_workflow', 'maa_ai_workflow'),
    (r'gva_codegen', 'maa_codegen'),
    (r'gva_polish', 'maa_polish'),
    (r"'gva:chunk", "'maa:chunk"),
    (r"dbName: 'gva'", "dbName: 'maa'"),
    (r'【不填则为GVA库】', '【不填则为默认库】'),
    (r'使用GVA结构', '使用默认结构'),
    (r'开启GVA默认结构', '开启默认结构'),
    (r"'GVA库'", "'默认库'"),
    (r'GVA 代码生成', '代码生成'), (r'GVA 功能完善', '功能完善'),
    (r'gva本库', '默认库'),
    (r'则默认使用gva库', '则默认使用默认库'),
    (r'^(\s*)gva: \{\s*$', r'\1maa: {'),
    (r'#gva-loading-box', '#maa-loading-box'),
    (r"reskins gin-vue-admin's stock system pages", 'reskins the inherited stock system pages'),
    (r'/\* ---- gva page containers ---- \*/', '/* ---- page containers ---- */'),
    (r'global\.GVA_', 'global.MAA_'),

    (r'// Author (SliverHorn|Kafumio|piexlmax)\s*$', '// Author MiTo'),
    (r'^\s*//\tAuthor \[SliverHorn\].*\n', ''),
    (r'\t\tif err := tx\.Model\(&system\.SysBaseMenu\{\}\)\.\n\t\t\tWhere\("name = \?", "https://www\.gin-vue-admin\.com"\)\.\n\t\t\tUpdates\(map\[string\]any\{\n(?:.+\n){4}\t\t\t\}\)\.Error; err != nil \{\n\t\t\treturn err\n\t\t\}\n', ''),
    (r'\t\t\{ParentId: 0, MenuLevel: 0, Path: "https://www\.gin-vue-admin\.com", Name: "https://www\.gin-vue-admin\.com"\},\n', ''),
    (r'\n\tvar projectHome system\.SysBaseMenu\n\tif err = db\.First\(&projectHome, legacyMenus\[8\]\.ID\)\.Error; err != nil \{\n\t\tt\.Fatalf\("reload migrated project home: %v", err\)\n\t\}\n\tif projectHome\.Name != "https://github\.com/WangMi2022/YuYuan-Pass-System" \|\| projectHome\.Path != projectHome\.Name \|\| !projectHome\.Hidden \{\n\t\tt\.Fatalf\("unexpected migrated project home: %#v", projectHome\)\n\t\}\n', '\n'),
    (r' \* @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: [^\n]*\n', ''),
    (r'<ai-gva\s*/>', '<ai-maa />'),
    (r'^(\s*)gva: \{\s*$', r'\1maa: {'),
    (r'\{Name: "10\.png", Url: "[^"]*gvalogo\.png"[^\n]*\n', ''),
]

# file renames baseline -> current
RENAMES = {
    'server/mcp/gva_execute.go': 'server/mcp/maa_execute.go',
    'server/mcp/gva_review.go': 'server/mcp/maa_review.go',
    'server/mcp/gva_analyze.go': 'server/mcp/maa_analyze.go',
    'web/src/assets/icons/ai-gva.svg': 'web/src/assets/icons/ai-maa.svg',
    'web/src/assets/icons/customer-gva.svg': 'web/src/assets/icons/customer-maa.svg',
}
DELETED = {'web/src/assets/flipped-aurora.png'}

COMPILED = [(re.compile(p), r) for p, r in RULES]

def normalize(text: str) -> str:
    for rx, rep in COMPILED:
        text = rx.sub(rep, text)
    return text

def walk(root, exts):
    for dp, dn, fn in os.walk(root):
        dn[:] = [d for d in dn if d not in ('node_modules', 'dist', '.deploy', 'uploads', 'log', 'docs') or (Path(dp).relative_to(ROOT).parts[:1] != ('server',) and d == 'docs')]
        for f in fn:
            p = Path(dp) / f
            if p.suffix in exts:
                yield p

def main():
    scopes = {
        'server': (('server',), {'.go'}),
        'web':    (('web', 'src'), {'.vue', '.js', '.scss', '.css', '.json', '.html'}),
        'deploy': (('deploy',), {'.sh', '.yml', '.yaml', '.py', '.example', '.Dockerfile'}),
    }
    total = explained = residual = 0
    residual_detail = []
    for name, (sub, exts) in scopes.items():
        base_root = BASE.joinpath(*sub)
        cur_root = ROOT.joinpath(*sub)
        for bp in walk(base_root, exts):
            rel = bp.relative_to(BASE).as_posix()
            if rel in DELETED: continue
            cur_rel = RENAMES.get(rel, rel)
            cp = ROOT / cur_rel
            if not cp.exists():
                residual_detail.append((rel, 'MISSING in current'))
                continue
            btext = bp.read_text(encoding='utf-8', errors='replace').replace('\r\n', '\n')
            ctext = cp.read_text(encoding='utf-8', errors='replace').replace('\r\n', '\n')
            if btext == ctext: continue
            total += 1
            expected = normalize(btext)
            if expected == ctext:
                explained += 1
                continue
            # show only the lines that differ
            d = list(difflib.unified_diff(expected.splitlines(), ctext.splitlines(), lineterm='', n=0))
            changes = [l for l in d if l.startswith(('+', '-')) and not l.startswith(('+++', '---'))]
            residual += 1
            residual_detail.append((cur_rel, changes[:8]))
    print(f"files changed since baseline : {total}")
    print(f"fully explained by rules     : {explained}")
    print(f"residual (needs explanation) : {residual}")
    print()
    for rel, ch in residual_detail:
        print(f"--- {rel}")
        if isinstance(ch, str):
            print(f"    {ch}")
        else:
            for l in ch: print(f"    {l[:160]}")

if __name__ == '__main__':
    main()
