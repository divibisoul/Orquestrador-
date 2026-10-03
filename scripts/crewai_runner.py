import json, os, sys
def emit(x,c=0): print(json.dumps(x,ensure_ascii=False)); raise SystemExit(c)
try:r=json.load(sys.stdin)
except Exception as e:emit({"state":"FAIL","code":"CREWAI_REQUEST_INVALID","detail":str(e)},2)
goal=str(r.get("goal") or "").strip(); root=os.path.abspath(str(r.get("root") or "integrations/external/crewai")); roles=r.get("roles") or ["Planner","Researcher"]
if not goal:emit({"state":"FAIL","code":"CREWAI_GOAL_REQUIRED"},2)
if not os.path.isdir(root):emit({"state":"DEGRADED","code":"CREWAI_SOURCE_NOT_AVAILABLE"})
if not os.environ.get("OPENAI_API_KEY"):emit({"state":"DEGRADED","code":"CREWAI_LLM_CREDENTIALS_NOT_AVAILABLE"})
\nfor _p in (root, os.path.join(root, "lib", "crewai", "src")):\n if os.path.isdir(_p) and _p not in sys.path: sys.path.insert(0, _p)
try:
 from crewai import Agent,Crew,LLM,Process,Task
except Exception as e:emit({"state":"DEGRADED","code":"CREWAI_PYTHON_IMPORT_FAILED","detail":str(e)})
try:
 llm=LLM(model=os.environ.get("SOUL_N07_CREWAI_MODEL","gpt-5-mini"))
 agents=[Agent(role=str(x),goal=goal,backstory="SOUL N07 specialist",llm=llm,allow_delegation=False,verbose=False) for x in roles]
 task=Task(description=goal,expected_output="Evidence-aware concise result.",agent=agents[0])
 out=Crew(agents=agents,tasks=[task],process=Process.sequential,verbose=False).kickoff()
except Exception as e:emit({"state":"FAIL","code":"CREWAI_EXECUTION_FAILED","detail":str(e)},2)
emit({"state":"PASS","provider":"crewai","goal":goal,"roles":[str(x) for x in roles],"result":str(out)})
