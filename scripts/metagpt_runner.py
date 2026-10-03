import asyncio,json,os,sys
def emit(x,c=0): print(json.dumps(x,ensure_ascii=False)); raise SystemExit(c)
try:r=json.load(sys.stdin)
except Exception as e:emit({"state":"FAIL","code":"METAGPT_REQUEST_INVALID","detail":str(e)},2)
goal=str(r.get("goal") or "").strip();root=os.path.abspath(str(r.get("root") or "integrations/external/metagpt"));rounds=max(1,min(20,int(r.get("maxRounds") or 3)))
if not goal:emit({"state":"FAIL","code":"METAGPT_GOAL_REQUIRED"},2)
if not os.path.isdir(root):emit({"state":"DEGRADED","code":"METAGPT_SOURCE_NOT_AVAILABLE"})
if not os.environ.get("OPENAI_API_KEY"):emit({"state":"DEGRADED","code":"METAGPT_LLM_CREDENTIALS_NOT_AVAILABLE"})
sys.path.insert(0,root)
try:from metagpt.team import Team
except Exception as e:emit({"state":"DEGRADED","code":"METAGPT_PYTHON_IMPORT_FAILED","detail":str(e)})
async def main():
 try:
  team=Team();team.invest(float(os.environ.get("SOUL_N07_METAGPT_INVESTMENT","10")));team.run_project(goal);h=await team.run(n_round=rounds,idea="",auto_archive=True);return {"state":"PASS","provider":"metagpt","goal":goal,"maxRounds":rounds,"history":str(h)}
 except Exception as e:return {"state":"FAIL","code":"METAGPT_EXECUTION_FAILED","detail":str(e)}
emit(asyncio.run(main()))
