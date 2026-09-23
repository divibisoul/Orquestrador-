/**
 * L6 — OMNI MODE MODULE
 * Host: N07 | Affinity: M7_EVOLUTION
 *
 * Coordination state only. "Omni" means the host has declared a mode in which
 * multiple operations may be tracked concurrently; it does not assert hardware
 * overclocking, unlimited throughput, or simultaneous execution by itself.
 */
export interface OmniOperation {
  id: string;
  startedAt: number;
}

export interface OmniModeState {
  enabled: boolean;
  activeOperations: readonly OmniOperation[];
  timestamp: number;
}

export class OmniModeModule {
  readonly id = "omni-mode" as const;
  private enabled = false;
  private readonly operations = new Map<string, OmniOperation>();

  enable(): OmniModeState {
    this.enabled = true;
    return this.getState();
  }

  disable(): OmniModeState {
    this.enabled = false;
    this.operations.clear();
    return this.getState();
  }

  startOperation(operationId: string): void {
    if (!this.enabled) return;
    const id = operationId.trim();
    if (!id) throw new Error("OMNI_OPERATION_ID_REQUIRED");
    if (!this.operations.has(id)) {
      this.operations.set(id, { id, startedAt: Date.now() });
    }
  }

  endOperation(operationId: string): void {
    this.operations.delete(operationId);
  }

  getActiveOperations(): OmniOperation[] {
    return [...this.operations.values()].map(item => ({ ...item }));
  }

  isEnabled(): boolean {
    return this.enabled;
  }

  getState(): OmniModeState {
    return {
      enabled: this.enabled,
      activeOperations: this.getActiveOperations(),
      timestamp: Date.now(),
    };
  }
}

export const omniModeModule = new OmniModeModule();
