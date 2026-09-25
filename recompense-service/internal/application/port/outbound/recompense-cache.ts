import type { Recompense } from "../../../domain/recompense";

export interface RecompenseCache {
    get(id: string): Promise<Recompense | undefined>;
    set(id: string, r: Recompense): Promise<void>;
    del(id: string): Promise<void>;
}
