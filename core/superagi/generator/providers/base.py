from abc import ABC, abstractmethod
from typing import Any


class BaseProvider(ABC):
    """Common interface for SuperAGI generation providers."""

    @abstractmethod
    def generate(self, prompt: str, **kwargs: Any) -> str:
        """Generate text from a prompt."""
        raise NotImplementedError
