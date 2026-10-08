#!/usr/bin/env python3
import subprocess, re, sys, os
P=os.environ['PGURL']
objs=sys.argv[1:] or ['app.orders','app.order_line','app.part','app.part_eu','app.rng','app.base','app.child1','app.child2','app.unl','app.rls','app.v_acct','app.v_secure','app.v_check','app.mv','app.sq','app.sq_owned','app.ft','app.account','app.account_partial','app.account_lower','app.account_meta_gin','app.mood','app.pair','app.noowner','"Mixed Case"."Odd Table"']
def run(tool,cmd):
    if tool=='psql': a=['psql',P,'-X','-c',cmd]
    else: a=[os.environ['USQL_'+tool.upper()],P,'-c',cmd]
    return subprocess.run(a,capture_output=True,text=True,timeout=30).stdout.splitlines()
def sections(lines):
    secs=[]
    for l in lines:
        if l and not l.startswith(' ') and not l.startswith('-') and not l.startswith('(') :
            m=re.match(r'^([A-Za-z][A-Za-z \-?]*?)(:|\s\()',l)
            secs.append(m.group(1).strip() if m else l[:30])
    return secs
for o in objs:
    p=sections(run('psql','\\d+ '+o)); n=sections(run('new','\\d+ '+o)); 
    miss=[s for s in p if s not in n]
    print(f'{o}: psql={p}\n    missing-in-new={miss} extra-in-new={[s for s in n if s not in p]}')
