#!/usr/bin/env python3
"""For columns that psql and the new usql both have, list the cells that differ."""
import os, sys
sys.argv=[sys.argv[0]]
exec(open(os.path.dirname(os.path.abspath(__file__))+'/compare.py').read().split('report = []')[0])
out=[]
for f in sorted(os.listdir(D)):
    if not f.startswith('cmd.'): continue
    n=f[4:]; cmd=open(f'{D}/{f}').read().strip()
    ph,pb,_=parse(load('psql',n)); nh,nb,_=parse(load('new',n))
    if not ph or not nh or ph[0]=='ERR' or nh[0]=='ERR': continue
    from collections import Counter
    def kk(h,r): return tuple(r[:2])+((r[h.index('Argument data types')],) if 'Argument data types' in h else ())
    pc=Counter(kk(ph,r) for r in pb); nc=Counter(kk(nh,r) for r in nb)
    pk={kk(ph,r):r for r in pb if pc[kk(ph,r)]==1}; nk={kk(nh,r):r for r in nb if nc[kk(nh,r)]==1}
    diffs={}
    for k,pr in pk.items():
        nr=nk.get(k)
        if not nr: continue
        for c in ph:
            if c in nh:
                a=pr[ph.index(c)]; b=nr[nh.index(c)]
                if a.replace('<NULL>','')!=b.replace('<NULL>',''): diffs.setdefault(c,[]).append((k,a,b))
    for c,l in diffs.items():
        out.append(f'`{cmd}` column {c}: {len(l)} differ, e.g. {l[0][0]}: psql={l[0][1][:70]!r} new={l[0][2][:70]!r}')
print('\n'.join(out))
