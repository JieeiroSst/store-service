import type { MemberStatus } from "../../../domain/loyalty";
import type { Recompense } from "../../../domain/recompense";
import type { EarnResult } from "../outbound/points-ledger";

export interface Redemption {
    Recompense: Recompense;
    Amount: number;
    Discount: number;
    Total: number;
}

export interface LoyaltyService {
    earn(memberId: number, idempotencyKey: string, amount: number): Promise<EarnResult & { status: MemberStatus }>;
    status(memberId: number): Promise<MemberStatus>;
    best(memberId: number, amount: number): Promise<Redemption | undefined>;
    redeem(recompenseId: string, memberId: number, amount: number): Promise<Redemption | undefined>;
}
