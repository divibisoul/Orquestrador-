from __future__ import annotations

from dataclasses import dataclass, field
from typing import Callable, Dict, List, Optional


@dataclass(frozen=True)
class EvolutionProposal:
    """A proposed change; proposals are inert until explicitly approved."""

    identifier: str
    description: str
    validator: Optional[Callable[[], bool]] = None
    metadata: Dict[str, str] = field(default_factory=dict)


class SelfEvolution:
    """Safe proposal registry for future self-evolution workflows.

    This component never mutates source code by itself. A proposal must be
    registered, validated, and explicitly approved by an external controller.
    """

    def __init__(self) -> None:
        self._proposals: Dict[str, EvolutionProposal] = {}
        self._approved: List[str] = []

    def propose(self, proposal: EvolutionProposal) -> None:
        if not proposal.identifier.strip():
            raise ValueError("proposal identifier is required")
        if proposal.identifier in self._proposals:
            raise ValueError("proposal already exists")
        self._proposals[proposal.identifier] = proposal

    def validate(self, identifier: str) -> bool:
        proposal = self._proposals.get(identifier)
        if proposal is None:
            raise KeyError(identifier)
        return True if proposal.validator is None else bool(proposal.validator())

    def approve(self, identifier: str) -> None:
        if not self.validate(identifier):
            raise ValueError("proposal validation failed")
        if identifier not in self._approved:
            self._approved.append(identifier)

    def get(self, identifier: str) -> EvolutionProposal:
        return self._proposals[identifier]

    def approved(self) -> List[str]:
        return list(self._approved)
