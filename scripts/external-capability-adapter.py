#!/usr/bin/env python3
import importlib.util, json, os, pathlib, subprocess, sys
MAX_INPUT=131072
MAX_OUTPUT=262144
PACKAGES={"superagi":"superagi","langgraph":"langgraph","crewai":"crewai","microsoft-agent-framework":"agent_framework","openhands":"openhands","metagpt":"metagpt","agentscope":"agentscope","browser-use":"browser_use","smolagents":"smolagents","pydantic-ai":"pydantic_ai","llama-index":"llama_index","dspy":"dspy","whisper":"whisper","kokoro":"kokoro"}

def emit(v,c=0):
    print(json.dumps(v,ensure_ascii=False)[:MAX_OUTPUT])
    raise SystemExit(c)

def load():
    raw=sys.stdin.read(MAX_INPUT+1)
    if len(raw)>MAX_INPUT: emit({"state":"BLOCKED","code":"EXTERNAL_ADAPTER_INPUT_TOO_LARGE"},2)
    try: v=json.loads(raw)
    except Exception as e: emit({"state":"BLOCKED","code":"EXTERNAL_ADAPTER_INVALID_JSON","detail":str(e)},2)
    if not isinstance(v,dict): emit({"state":"BLOCKED","code":"EXTERNAL_ADAPTER_REQUEST_NOT_OBJECT"},2)
    return v

def root(v):
    p=pathlib.Path(str(v.get("root") or "")).resolve()
    if not p.exists(): emit({"state":"BLOCKED","code":"EXTERNAL_SOURCE_NOT_PRESENT","root":str(p)},2)
    return p

def package_available(p):
    name=PACKAGES.get(p)
    return name is not None and importlib.util.find_spec(name) is not None

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
        run_script("crewai_runner.py",{"goal":v.get("goal") or v.get("prompt"),"roles":v.get("roles") or ["Planner","Executor"],"root":str(rt),"maxRounds":v.get("maxRounds") or 3})
    if provider=="metagpt" and op=="team.execute":
        run_script("metagpt_runner.py",{"goal":v.get("goal") or v.get("prompt"),"root":str(rt),"maxRounds":v.get("maxRounds") or 3})
    if provider=="whisper" and op=="speech.transcribe":
        audio=pathlib.Path(str(v.get("audio_path") or "")).resolve()
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
    if provider=="kokoro" and op=="speech.synthesize":
        text=str(v.get("text") or "").strip()
        if not text: emit({"state":"BLOCKED","code":"KOKORO_TEXT_REQUIRED"},2)
        out=pathlib.Path(str(v.get("output_path") or "artifacts/kokoro-output.wav")).resolve()
        try:
            from kokoro import KPipeline
            import numpy as np, soundfile as sf
            pipe=KPipeline(lang_code=str(v.get("lang") or "a"))
            chunks=[a for _,_,a in pipe(text,voice=str(v.get("voice") or "af_heart"))]
            if not chunks: raise RuntimeError("KOKORO_NO_AUDIO")
            out.parent.mkdir(parents=True,exist_ok=True); sf.write(str(out),np.concatenate(chunks),24000)
            emit({"state":"PASS","provider":provider,"operation":op,"output_path":str(out)})
        except Exception as e: emit({"state":"BLOCKED","code":"KOKORO_EXECUTION_BLOCKED","detail":str(e)},2)
    if op in {"agent.describe","workflow.describe","browser.describe"}:
        emit({"state":"PASS","provider":provider,"operation":op,"source_present":rt.exists(),"package_available":package_available(provider)})
    emit({"state":"BLOCKED","code":"EXTERNAL_ADAPTER_OPERATION_NOT_AVAILABLE","provider":provider,"operation":op},2)

v=load()
rt=root(v)
if str(v.get("mode") or "probe").strip().lower()=="probe":
    emit({"state":"PASS","provider":str(v.get("provider") or ""),"mode":"probe","source_present":rt.exists(),"package_available":package_available(str(v.get("provider") or "")),"root":str(rt)})
execute(v,rt)
