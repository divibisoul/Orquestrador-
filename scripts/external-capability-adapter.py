#!/usr/bin/env python3
import importlib.util, json, os, pathlib, subprocess, sys, time
MAX_INPUT=131072
MAX_OUTPUT=262144
REPO_ROOT=pathlib.Path(__file__).resolve().parents[1]
PACKAGES={"fedml":"fedml","hivemind":"hivemind","ray":"ray","superagi":"superagi","langgraph":"langgraph","crewai":"crewai","microsoft-agent-framework":"agent_framework","openhands":"openhands","metagpt":"metagpt","agentscope":"agentscope","browser-use":"browser_use","smolagents":"smolagents","pydantic-ai":"pydantic_ai","llama-index":"llama_index","dspy":"dspy","whisper":"whisper","kokoro":"kokoro"}

def emit(v,c=0):
    raw=json.dumps(v,ensure_ascii=False)
    if len(raw)>MAX_OUTPUT:
        print(json.dumps({"state":"BLOCKED","code":"EXTERNAL_ADAPTER_OUTPUT_TOO_LARGE"},ensure_ascii=False))
        raise SystemExit(2)
    print(raw)
    raise SystemExit(c)

def load():
    raw=sys.stdin.read(MAX_INPUT+1)
    if len(raw)>MAX_INPUT: emit({"state":"BLOCKED","code":"EXTERNAL_ADAPTER_INPUT_TOO_LARGE"},2)
    try: v=json.loads(raw)
    except Exception as e: emit({"state":"BLOCKED","code":"EXTERNAL_ADAPTER_INVALID_JSON","detail":str(e)},2)
    if not isinstance(v,dict): emit({"state":"BLOCKED","code":"EXTERNAL_ADAPTER_REQUEST_NOT_OBJECT"},2)
    return v

def root(v):
    provider=str(v.get("provider") or "").strip().lower()
    configured=REPO_ROOT/"integrations/external-adapters.json"
    try:
        adapters=json.loads(configured.read_text(encoding="utf-8"))
    except Exception as e:
        emit({"state":"BLOCKED","code":"EXTERNAL_ADAPTER_REGISTRY_UNREADABLE","detail":str(e)},2)
    providers=[*adapters.get("providers",[]), *adapters.get("expansion_providers",[])]
    entry=next((p for p in providers if p.get("id")==provider),None)
    if not entry:
        emit({"state":"BLOCKED","code":"EXTERNAL_PROVIDER_NOT_REGISTERED","provider":provider},2)
    expected=(REPO_ROOT/str(entry["root"])).resolve()
    requested=pathlib.Path(str(v.get("root") or entry["root"]))
    if not requested.is_absolute():
        requested=REPO_ROOT/requested
    p=requested.resolve()
    try:
        p.relative_to(REPO_ROOT)
    except ValueError:
        emit({"state":"BLOCKED","code":"EXTERNAL_ROOT_ESCAPE","provider":provider,"root":str(p)},2)
    if p != expected:
        emit({"state":"BLOCKED","code":"EXTERNAL_ROOT_NOT_CANONICAL","provider":provider,"expected":str(expected),"requested":str(p)},2)
    if not p.exists() and str(v.get("mode") or "").strip().lower() != "probe":
        emit({"state":"BLOCKED","code":"EXTERNAL_SOURCE_NOT_PRESENT","root":str(p)},2)
    return p

def package_available(p):
    name=PACKAGES.get(p)
    return name is not None and importlib.util.find_spec(name) is not None

def external_io_root():
    configured=str(os.environ.get("SOUL_EXTERNAL_IO_ROOT") or "").strip()
    if configured:
        p=pathlib.Path(configured).resolve()
    else:
        p=(REPO_ROOT/"artifacts/external-io").resolve()
    p.mkdir(parents=True,exist_ok=True)
    try: p.relative_to(REPO_ROOT)
    except ValueError:
        if configured:
            return p
    return p

def external_io_path(raw, default_name):
    root=external_io_root()
    value=str(raw or default_name).strip()
    p=pathlib.Path(value)
    if not p.is_absolute():
        p=root/p
    p=p.resolve()
    try:
        p.relative_to(root)
    except ValueError:
        emit({"state":"BLOCKED","code":"EXTERNAL_IO_PATH_ESCAPE","root":str(root),"path":str(p)},2)
    return p

def read_skill(rt,v):
    rel=str(v.get("path") or "README.md").strip()
    p=(rt/rel).resolve()
    try: p.relative_to(rt)
    except ValueError: emit({"state":"BLOCKED","code":"SUPERPOWERS_PATH_ESCAPE"},2)
    if p.suffix.lower()!=".md" or not p.is_file(): emit({"state":"BLOCKED","code":"SUPERPOWERS_SKILL_NOT_FOUND","path":rel},2)
    emit({"state":"PASS","provider":"superpowers","operation":"skills.read","path":rel,"content":p.read_text(encoding="utf-8")[:65536]})

def run_script(name,payload):
    script=pathlib.Path(__file__).with_name(name)
    p=subprocess.run([sys.executable,str(script)],input=json.dumps(payload),text=True,capture_output=True,timeout=170,check=False,env=os.environ.copy())
    if p.returncode!=0: emit({"state":"BLOCKED","code":name.upper().replace(".","_")+"_FAILED","detail":p.stderr[-4000:]},2)
    try: result=json.loads(p.stdout)
    except Exception as e: emit({"state":"BLOCKED","code":"EXTERNAL_RUNNER_INVALID_OUTPUT","detail":str(e)},2)
    emit(result,0 if result.get("state")=="PASS" else 2)


def execute_agentscope(v,rt):
    prompt=str(v.get("prompt") or v.get("input") or "").strip()
    if not prompt: emit({"state":"BLOCKED","code":"AGENTSCOPE_PROMPT_REQUIRED"},2)
    if not os.environ.get("OPENAI_API_KEY"): emit({"state":"BLOCKED","code":"AGENTSCOPE_OPENAI_CREDENTIALS_REQUIRED"},2)
    try:
        import asyncio
        from agentscope.agent import Agent
        from agentscope.credential import OpenAICredential
        from agentscope.message import UserMsg
        from agentscope.model import OpenAIChatModel
        async def run():
            agent=Agent(
                name=str(v.get("name") or "SOUL-AgentScope"),
                system_prompt=str(v.get("system") or "You are an SOUL augmentation agent."),
                model=OpenAIChatModel(
                    credential=OpenAICredential(api_key=os.environ["OPENAI_API_KEY"]),
                    model=str(v.get("model") or os.environ.get("SOUL_EXTERNAL_AGENTSCOPE_MODEL") or "gpt-5-mini"),
                    stream=False,
                ),
            )
            reply=await agent.reply(UserMsg(name="user",content=prompt))
            return reply.get_text_content()
        emit({"state":"PASS","provider":"agentscope","operation":"agent.execute","result":asyncio.run(run())})
    except Exception as e: emit({"state":"BLOCKED","code":"AGENTSCOPE_EXECUTION_BLOCKED","detail":str(e)},2)

def execute_browser(v,rt):
    task=str(v.get("task") or v.get("prompt") or "").strip()
    if not task: emit({"state":"BLOCKED","code":"BROWSER_USE_TASK_REQUIRED"},2)
    if not os.environ.get("BROWSER_USE_API_KEY") and not os.environ.get("OPENAI_API_KEY"):
        emit({"state":"BLOCKED","code":"BROWSER_USE_CREDENTIALS_REQUIRED"},2)
    try:
        import asyncio
        from browser_use import Agent, ChatBrowserUse
        async def run():
            agent=Agent(task=task,llm=ChatBrowserUse(model=v.get("model") or None) if v.get("model") else ChatBrowserUse())
            result=await agent.run(max_steps=max(1,min(500,int(v.get("max_steps") or 50))))
            final=getattr(result,"final_result",None)
            return final() if callable(final) else str(result)
        emit({"state":"PASS","provider":"browser-use","operation":"browser.execute","result":asyncio.run(run())})
    except Exception as e: emit({"state":"BLOCKED","code":"BROWSER_USE_EXECUTION_BLOCKED","detail":str(e)},2)

def execute_smolagents(v,rt):
    prompt=str(v.get("prompt") or v.get("input") or "").strip()
    if not prompt: emit({"state":"BLOCKED","code":"SMOLAGENTS_PROMPT_REQUIRED"},2)
    if not os.environ.get("OPENAI_API_KEY"): emit({"state":"BLOCKED","code":"SMOLAGENTS_OPENAI_CREDENTIALS_REQUIRED"},2)
    try:
        from smolagents import OpenAIModel, ToolCallingAgent
        model=OpenAIModel(model_id=str(v.get("model") or os.environ.get("SOUL_EXTERNAL_SMOLAGENTS_MODEL") or "gpt-5-mini"))
        agent=ToolCallingAgent(tools=[],model=model,add_base_tools=False)
        emit({"state":"PASS","provider":"smolagents","operation":"agent.execute","result":str(agent.run(prompt))})
    except Exception as e: emit({"state":"BLOCKED","code":"SMOLAGENTS_EXECUTION_BLOCKED","detail":str(e)},2)

def execute_microsoft_agent_framework(v,rt):
    prompt=str(v.get("prompt") or v.get("input") or "").strip()
    if not prompt: emit({"state":"BLOCKED","code":"MICROSOFT_AGENT_FRAMEWORK_PROMPT_REQUIRED"},2)
    if not os.environ.get("OPENAI_API_KEY"): emit({"state":"BLOCKED","code":"MICROSOFT_AGENT_FRAMEWORK_OPENAI_CREDENTIALS_REQUIRED"},2)
    try:
        import asyncio
        from agent_framework import Agent
        from agent_framework.openai import OpenAIChatClient
        async def run():
            client=OpenAIChatClient(model=str(v.get("model") or os.environ.get("SOUL_EXTERNAL_MAF_MODEL") or "gpt-5-mini"),api_key=os.environ["OPENAI_API_KEY"])
            agent=Agent(client=client,name=str(v.get("name") or "SOUL-MAF"),instructions=str(v.get("system") or "You are an SOUL augmentation agent."))
            response=await agent.run(prompt)
            return getattr(response,"text",str(response))
        emit({"state":"PASS","provider":"microsoft-agent-framework","operation":"agent.execute","result":asyncio.run(run())})
    except Exception as e: emit({"state":"BLOCKED","code":"MICROSOFT_AGENT_FRAMEWORK_EXECUTION_BLOCKED","detail":str(e)},2)

def execute_openhands(v,rt):
    import urllib.request
    configured_base=str(os.environ.get("SOUL_EXTERNAL_OPENHANDS_URL") or "").strip().rstrip("/")
    requested_base=str(v.get("base_url") or "").strip().rstrip("/")
    if requested_base and requested_base!=configured_base:
        emit({"state":"BLOCKED","code":"OPENHANDS_BASE_URL_OVERRIDE_FORBIDDEN"},2)
    base=configured_base
    conversation=str(v.get("conversation_id") or os.environ.get("SOUL_EXTERNAL_OPENHANDS_CONVERSATION_ID") or "").strip()
    prompt=str(v.get("prompt") or v.get("input") or "").strip()
    if not base or not conversation or not prompt:
        emit({"state":"BLOCKED","code":"OPENHANDS_BASE_URL_CONVERSATION_AND_PROMPT_REQUIRED"},2)
    path=f"{base}/api/conversations/{conversation}/ask_agent"
    body=json.dumps({"question":prompt}).encode()
    req=urllib.request.Request(path,data=body,headers={"Content-Type":"application/json"})
    if os.environ.get("SOUL_EXTERNAL_OPENHANDS_SESSION_KEY"):
        req.add_header("X-Session-API-Key",os.environ["SOUL_EXTERNAL_OPENHANDS_SESSION_KEY"])
    try:
        with urllib.request.urlopen(req,timeout=min(120,int(v.get("timeout_seconds") or 60))) as resp:
            result=json.loads(resp.read(MAX_OUTPUT))
        emit({"state":"PASS","provider":"openhands","operation":"agent.execute","result":result})
    except Exception as e: emit({"state":"BLOCKED","code":"OPENHANDS_EXECUTION_BLOCKED","detail":str(e)},2)

def execute_letta(v,rt):
    agent=str(v.get("agent_id") or "").strip()
    prompt=str(v.get("prompt") or v.get("input") or "").strip()
    if not prompt: emit({"state":"BLOCKED","code":"LETTA_PROMPT_REQUIRED"},2)
    binary=str(os.environ.get("SOUL_EXTERNAL_LETTA_BIN") or "letta")
    args=[binary,"-p"]
    if agent: args+=["--agent",agent]
    else: args+=["--new-agent"]
    args += [prompt]
    try:
        p=subprocess.run(args,text=True,capture_output=True,timeout=min(600,int(v.get("timeout_seconds") or 120)),env=os.environ.copy(),check=False)
        if p.returncode!=0: emit({"state":"BLOCKED","code":"LETTA_EXECUTION_BLOCKED","detail":p.stderr[-4000:]},2)
        emit({"state":"PASS","provider":"letta-code","operation":"memory.execute","result":p.stdout[:MAX_OUTPUT]})
    except Exception as e: emit({"state":"BLOCKED","code":"LETTA_PROCESS_FAILED","detail":str(e)},2)

def execute(v,rt):
    provider=str(v.get("provider") or "").strip().lower()
    op=str(v.get("operation") or "").strip().lower()
    if provider=="superpowers" and op=="skills.read": read_skill(rt,v)
    if provider=="langgraph" and op=="workflow.invoke":
        try:
            from langgraph.graph import END, START, StateGraph
            g=StateGraph(dict)
            g.add_node("soul",lambda s:{"soul_output":s.get("input")})
            g.add_edge(START,"soul"); g.add_edge("soul",END)
            emit({"state":"PASS","provider":provider,"operation":op,"result":g.compile().invoke({"input":v.get("input") or v.get("prompt") or ""})})
        except Exception as e: emit({"state":"BLOCKED","code":"LANGGRAPH_EXECUTION_BLOCKED","detail":str(e)},2)
    if provider=="crewai" and op=="team.execute":
        run_script("crewai_runner.py",{"goal":v.get("goal") or v.get("prompt"),"roles":v.get("roles") or ["Planner","Executor"],"maxRounds":v.get("maxRounds") or 3})
    if provider=="metagpt" and op=="team.execute":
        run_script("metagpt_runner.py",{"goal":v.get("goal") or v.get("prompt"),"maxRounds":v.get("maxRounds") or 3})
    if provider=="whisper" and op=="speech.transcribe":
        audio=external_io_path(v.get("audio_path"), "input.wav")
        if not audio.is_file(): emit({"state":"BLOCKED","code":"WHISPER_AUDIO_NOT_FOUND"},2)
        try:
            import whisper
            result=whisper.load_model(str(v.get("model") or "base")).transcribe(str(audio),fp16=False)
            emit({"state":"PASS","provider":provider,"operation":op,"text":result.get("text",""),"language":result.get("language")})
        except Exception as e: emit({"state":"BLOCKED","code":"WHISPER_EXECUTION_BLOCKED","detail":str(e)},2)
    if provider=="pydantic-ai" and op=="agent.execute":
        prompt=str(v.get("prompt") or v.get("input") or "").strip()
        if not prompt: emit({"state":"BLOCKED","code":"PYDANTIC_AI_PROMPT_REQUIRED"},2)
        try:
            from pydantic_ai import Agent
            model=str(v.get("model") or os.environ.get("SOUL_EXTERNAL_PYDANTIC_AI_MODEL") or "openai:gpt-5-mini")
            result=Agent(model).run_sync(prompt)
            emit({"state":"PASS","provider":provider,"operation":op,"result":str(getattr(result,"output",result))})
        except Exception as e: emit({"state":"BLOCKED","code":"PYDANTIC_AI_EXECUTION_BLOCKED","detail":str(e)},2)
    if provider=="llama-index" and op=="retrieval.execute":
        source=str(v.get("text") or v.get("document") or "").strip()
        query=str(v.get("query") or "").strip()
        if not source or not query: emit({"state":"BLOCKED","code":"LLAMA_INDEX_TEXT_AND_QUERY_REQUIRED"},2)
        try:
            from llama_index.core import Document,VectorStoreIndex
            idx=VectorStoreIndex.from_documents([Document(text=source)])
            nodes=idx.as_retriever(similarity_top_k=3).retrieve(query)
            emit({"state":"PASS","provider":provider,"operation":op,"matches":[{"text":getattr(n,"text",str(n)),"score":getattr(n,"score",None)} for n in nodes]})
        except Exception as e: emit({"state":"BLOCKED","code":"LLAMA_INDEX_EXECUTION_BLOCKED","detail":str(e)},2)
    if provider=="dspy" and op=="reasoning.execute":
        prompt=str(v.get("prompt") or v.get("input") or "").strip()
        if not prompt: emit({"state":"BLOCKED","code":"DSPY_PROMPT_REQUIRED"},2)
        try:
            import dspy
            dspy.configure(lm=dspy.LM(str(v.get("model") or os.environ.get("SOUL_EXTERNAL_DSPY_MODEL") or "openai/gpt-5-mini")))
            r=dspy.Predict("question -> answer")(question=prompt)
            emit({"state":"PASS","provider":provider,"operation":op,"result":str(getattr(r,"answer",r))})
        except Exception as e: emit({"state":"BLOCKED","code":"DSPY_EXECUTION_BLOCKED","detail":str(e)},2)
    if provider=="agentscope" and op=="agent.execute":
        execute_agentscope(v,rt)
    if provider=="browser-use" and op=="browser.execute":
        execute_browser(v,rt)
    if provider=="smolagents" and op=="agent.execute":
        execute_smolagents(v,rt)
    if provider=="microsoft-agent-framework" and op=="agent.execute":
        execute_microsoft_agent_framework(v,rt)
    if provider=="openhands" and op=="agent.execute":
        execute_openhands(v,rt)
    if provider=="letta-code" and op=="memory.execute":
        execute_letta(v,rt)
    if provider=="superagi" and op=="agent.execute":
        configured_base=str(os.environ.get("SOUL_EXTERNAL_SUPERAGI_URL") or "").strip().rstrip("/")
        requested_base=str(v.get("base_url") or "").strip().rstrip("/")
        if requested_base and requested_base!=configured_base:
            emit({"state":"BLOCKED","code":"SUPERAGI_BASE_URL_OVERRIDE_FORBIDDEN"},2)
        base=configured_base
        agent_id=str(v.get("agent_id") or os.environ.get("SOUL_EXTERNAL_SUPERAGI_AGENT_ID") or "").strip()
        api_key=str(os.environ.get("SOUL_EXTERNAL_SUPERAGI_API_KEY") or "").strip()
        goal=str(v.get("goal") or v.get("prompt") or "").strip()
        instruction=str(v.get("instruction") or "").strip()
        if not base or not agent_id or not api_key:
            emit({"state":"BLOCKED","code":"SUPERAGI_BASE_URL_AGENT_ID_AND_API_KEY_REQUIRED"},2)
        if not goal and not instruction:
            emit({"state":"BLOCKED","code":"SUPERAGI_GOAL_OR_INSTRUCTION_REQUIRED"},2)
        import urllib.request
        headers={"Content-Type":"application/json","X-API-Key":api_key}
        try:
            payload={}
            if goal: payload["goal"]=[goal]
            if instruction: payload["instruction"]=[instruction]
            req=urllib.request.Request(f"{base}/v1/agent/{agent_id}/run",data=json.dumps(payload).encode(),headers=headers)
            with urllib.request.urlopen(req,timeout=30) as resp:
                result=json.loads(resp.read(MAX_OUTPUT))
            run_id=result.get("run_id")
            if run_id is None:
                emit({"state":"BLOCKED","code":"SUPERAGI_RUN_ID_MISSING","result":result},2)
            deadline=time.monotonic()+min(120,int(v.get("timeout_seconds") or 60))
            last_status="RUNNING"
            while time.monotonic()<deadline:
                status_req=urllib.request.Request(
                    f"{base}/v1/agent/{agent_id}/run-status",
                    data=json.dumps({"run_ids":[int(run_id)],"run_status_filter":None}).encode(),
                    headers=headers,
                )
                with urllib.request.urlopen(status_req,timeout=15) as resp:
                    status_result=json.loads(resp.read(MAX_OUTPUT))
                if isinstance(status_result,list) and status_result:
                    last_status=str(status_result[0].get("status") or "")
                elif isinstance(status_result,dict):
                    last_status=str(status_result.get("status") or "")
                if last_status.upper()=="COMPLETED":
                    emit({"state":"PASS","provider":provider,"operation":"agent.execute","run_id":int(run_id),"completion_proven":True,"status":last_status})
                if last_status.upper() in {"TERMINATED","PAUSED","WAITING_FOR_PERMISSION","WAIT_STEP","ITERATION_LIMIT_EXCEEDED"}:
                    emit({"state":"BLOCKED","code":"SUPERAGI_EXECUTION_NOT_COMPLETED","run_id":int(run_id),"status":last_status},2)
                time.sleep(1)
            emit({"state":"BLOCKED","code":"SUPERAGI_EXECUTION_TIMEOUT","run_id":int(run_id),"status":last_status},2)
        except Exception as ex:
            emit({"state":"BLOCKED","code":"SUPERAGI_EXECUTION_BLOCKED","detail":str(ex)},2)

    if provider=="kokoro" and op=="speech.synthesize":
        text=str(v.get("text") or "").strip()
        if not text: emit({"state":"BLOCKED","code":"KOKORO_TEXT_REQUIRED"},2)
        out=external_io_path(v.get("output_path"), "kokoro-output.wav")
        try:
            from kokoro import KPipeline
            import numpy as np, soundfile as sf
            pipe=KPipeline(lang_code=str(v.get("lang") or "a"))
            chunks=[a for _,_,a in pipe(text,voice=str(v.get("voice") or "af_heart"))]
            if not chunks: raise RuntimeError("KOKORO_NO_AUDIO")
            out.parent.mkdir(parents=True,exist_ok=True); sf.write(str(out),np.concatenate(chunks),24000)
            emit({"state":"PASS","provider":provider,"operation":op,"output_path":str(out)})
        except Exception as e: emit({"state":"BLOCKED","code":"KOKORO_EXECUTION_BLOCKED","detail":str(e)},2)
    if provider in {"fedml","hivemind","temporal","cognitive-workspace","ravana","ray"} and op not in {
        "federated.describe","swarm.describe","workflow.describe","workspace.describe","agent.describe","cluster.describe"
    }:
        emit({
            "state":"BLOCKED",
            "code":"SOUL_V2_EXPANSION_RUNTIME_NOT_BOUND",
            "provider":provider,
            "operation":op,
            "reason":"native adapter and observed runtime are required before execution promotion",
            "epistemic_state":"PROJECTED"
        },2)

    if op.endswith(".describe"):
        emit({
            "state":"REAL" if rt.exists() else "BLOCKED",
            "provider":provider,
            "operation":op,
            "source_present":rt.exists(),
            "package_available":package_available(provider),
            "execution_proven":False,
        })
    emit({"state":"BLOCKED","code":"EXTERNAL_ADAPTER_OPERATION_NOT_AVAILABLE","provider":provider,"operation":op},2)

v=load()
rt=root(v)
if str(v.get("mode") or "probe").strip().lower()=="probe":
    emit({"state":"PASS","provider":str(v.get("provider") or ""),"mode":"probe","source_present":rt.exists(),"package_available":package_available(str(v.get("provider") or "")),"root":str(rt)})
execute(v,rt)
