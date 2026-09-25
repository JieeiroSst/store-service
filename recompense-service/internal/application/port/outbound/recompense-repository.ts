import type { Recompense } from "../../../domain/recompense";

export interface RecompenseRepository {
    insert(r: Recompense): Promise<Recompense>;
    update(id: string, r: Recompense): Promise<Recompense | undefined>;
    getByID(id: string): Promise<Recompense | undefined>;
    getByMemberID(memberId: number, limit: number, offset: number): Promise<Recompense[]>;
    delete(id: string): Promise<boolean>;
    list(limit: number, offset: number): Promise<Recompense[]>;
    ping(): Promise<void>;
}
