#!/usr/bin/env python3
"""Compare the csv output of psql, the old usql and the new usql, per command."""
import csv, io, os, re, sys
D = os.path.dirname(os.path.abspath(__file__)) + '/out'
def load(tool, n):
    s = open(f'{D}/{tool}/{n}.csv', errors='replace').read()
    lines = s.splitlines()
    while lines and (lines[0].startswith('Output format') or lines[0].startswith('Null display')):
        lines.pop(0)
    return lines
def parse(lines):
    """first table: header, rows; and the section titles that follow."""
    if not lines: return None, [], []
    text = '\n'.join(lines)
    err = lines[0].startswith(('ERROR', 'error', 'Invalid command', 'psql:', 'unknown', 'Did not find'))
    if err: return ('ERR', lines[0][:100]), [], []
    rd = csv.reader(io.StringIO(text))
    rows = []
    try:
        for r in rd: rows.append(r)
    except Exception: pass
    if not rows: return None, [], []
    hdr = rows[0]
    body = []
    secs = []
    for r in rows[1:]:
        if len(r) == len(hdr): body.append(r)
        else: secs.append(' '.join(r)[:60])
    return hdr, body, secs
def key(hdr, row):
    return tuple(row[:2]) if hdr and len(row) >= 2 else tuple(row)
report = []
summary = []
n = 0
for f in sorted(os.listdir(D)):
    if not f.startswith('cmd.'): continue
    n = f[4:]
    cmd = open(f'{D}/{f}').read().strip()
    res = {t: parse(load(t, n)) for t in ('psql', 'old', 'new')}
    ph, pb, ps = res['psql']; oh, ob, os_ = res['old']; nh, nb, ns = res['new']
    out = [f'### `{cmd}`']
    def fmt(h): return ('ERROR: ' + h[1]) if h and h[0] == 'ERR' else (', '.join(h) if h else '(nothing)')
    out.append(f'- psql: {fmt(ph)} [{len(pb)} rows]')
    out.append(f'- old:  {fmt(oh)} [{len(ob)} rows]')
    out.append(f'- new:  {fmt(nh)} [{len(nb)} rows]')
    miss = []
    if ph and ph[0] != 'ERR' and nh and nh[0] != 'ERR':
        miss = [c for c in ph if c not in nh]
        out.append(f'- **missing vs psql**: {", ".join(miss) or "none"}')
        extra = [c for c in nh if c not in ph]
        if extra: out.append(f'- extra vs psql: {", ".join(extra)}')
        # row sets
        pk = {key(ph, r) for r in pb}; nk = {key(nh, r) for r in nb}
        if pk != nk:
            out.append(f'- rows only in psql: {sorted(pk - nk)[:6]}{" ..." if len(pk-nk)>6 else ""}')
            out.append(f'- rows only in new: {sorted(nk - pk)[:6]}{" ..." if len(nk-pk)>6 else ""}')
        if ps != ns and (ps or ns):
            out.append(f'- sections psql: {ps[:40]}')
            out.append(f'- sections new:  {ns[:40]}')
    elif ph and ph[0] != 'ERR' and nh and nh[0] == 'ERR':
        out.append('- **new fails where psql works**')
    if oh and oh[0] != 'ERR' and nh and nh[0] != 'ERR':
        om = [c for c in oh if c not in nh and c not in (ph or [])]
        if om: out.append(f'- old columns not in new (and not psql): {", ".join(om)}')
    report.append('\n'.join(out))
    status = 'FAIL' if (nh and nh[0] == 'ERR' and ph and ph[0] != 'ERR') else ('GAP' if miss else 'ok')
    summary.append(f'{status:5} {cmd:28} missing: {", ".join(miss)}')
open(os.path.dirname(D) + '/report.md', 'w').write('\n\n'.join(report) + '\n')
print('\n'.join(summary))
