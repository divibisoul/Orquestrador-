# NVIDIA NeMo Agent Toolkit adapter

This directory contains an optional observation/evaluation environment for SOUL N07 orchestration. It deliberately does not alter SARA ARA/ETR/ITR and does not create a Mesh identity or public endpoint.

The stable toolkit line is 1.8.x in the official documentation. The documented package is nvidia-nat, and the langchain extra includes LangChain/LangGraph plugins. The toolkit is designed to work alongside existing agentic frameworks rather than requiring a replatform.

Install:

python -m pip install -r integrations/nemo-agent-toolkit/requirements.txt

Verify:

nat --version

Real workflow/evaluation execution remains gated on model credentials and an explicit N07 integration target. No credentials are stored in this repository.
