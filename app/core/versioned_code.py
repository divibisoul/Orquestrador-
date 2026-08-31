from __future__ import annotations

from dataclasses import dataclass
from hashlib import sha256
from typing import Dict, List


@dataclass(frozen=True)
class CodeVersion:
    version: str
    digest: str
    source: str


class VersionedCode:
    """Immutable in-memory code-version registry.

    It records content and integrity digests but does not execute generated
    source. Execution belongs behind an explicit sandbox boundary.
    """

    def __init__(self) -> None:
        self._versions: Dict[str, CodeVersion] = {}

    @staticmethod
    def digest(source: str) -> str:
        return sha256(source.encode("utf-8")).hexdigest()

    def register(self, version: str, source: str) -> CodeVersion:
        if not version.strip():
            raise ValueError("version is required")
        if version in self._versions:
            raise ValueError("version already exists")
        item = CodeVersion(version=version, digest=self.digest(source), source=source)
        self._versions[version] = item
        return item

    def get(self, version: str) -> CodeVersion:
        return self._versions[version]

    def verify(self, version: str, source: str) -> bool:
        item = self.get(version)
        return item.digest == self.digest(source)

    def versions(self) -> List[str]:
        return list(self._versions)
