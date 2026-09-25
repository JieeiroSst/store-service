import type { Recompense } from "../../../domain/recompense";

export interface RecompenseService {
    create(input: Recompense): Promise<Recompense>;
    update(id: string, input: Recompense): Promise<Recompense | undefined>;
    get(id: string): Promise<Recompense | undefined>;
    listByMember(memberId: number, page: number, limit: number): Promise<Recompense[]>;
    delete(id: string): Promise<boolean>;
    list(page: number, limit: number): Promise<Recompense[]>;
    ready(): Promise<void>;
}
