import type { LoyaltyService, Redemption } from "../port/inbound/loyalty";
import type { PointsLedger } from "../port/outbound/points-ledger";
import type { RecompenseRepository } from "../port/outbound/recompense-repository";
import type { IdGenerator } from "../port/outbound/id-generator";
import { discountFor, earnPoints, ineligibleReason, memberStatus, selectBest } from "../../domain/loyalty";
import { NotEligibleError, ValidationError } from "../../domain/recompense";

const candidateLimit = 500;

function checkMember(memberId: number): void {
    if (!Number.isInteger(memberId) || memberId <= 0) throw new ValidationError("member id must be a positive integer");
}

function checkAmount(amount: number): void {
    if (!Number.isFinite(amount) || amount <= 0) throw new ValidationError("amount must be > 0");
}

export class LoyaltyServiceImpl implements LoyaltyService {
    constructor(
        private readonly repo: RecompenseRepository,
        private readonly ledger: PointsLedger,
        private readonly ids: IdGenerator,
        private readonly now: () => number = Date.now,
    ) { }

    async status(memberId: number) {
        checkMember(memberId);
        return memberStatus(memberId, await this.ledger.balance(memberId));
    }

    async earn(memberId: number, idempotencyKey: string, amount: number) {
        checkMember(memberId);
        checkAmount(amount);
        if (!idempotencyKey) throw new ValidationError("idempotencyKey is required");
        const before = memberStatus(memberId, await this.ledger.balance(memberId));
        const result = await this.ledger.earn(this.ids.nextId(), memberId, idempotencyKey, earnPoints(amount, before.Tier));
        return { ...result, status: memberStatus(memberId, result.balance) };
    }

    async best(memberId: number, amount: number): Promise<Redemption | undefined> {
        checkAmount(amount);
        const status = await this.status(memberId);
        const candidates = await this.repo.getByMemberID(memberId, candidateLimit, 0);
        const best = selectBest(candidates, status, amount, this.now());
        return best && { Recompense: best.recompense, Amount: amount, Discount: best.discount, Total: amount - best.discount };
    }

    async redeem(recompenseId: string, memberId: number, amount: number): Promise<Redemption | undefined> {
        checkAmount(amount);
        const recompense = await this.repo.getByID(recompenseId);
        if (!recompense) return undefined;
        if (recompense.MemberID !== memberId) throw new NotEligibleError("recompense belongs to another member");
        const status = await this.status(memberId);
        const reason = ineligibleReason(recompense, status, this.now());
        if (reason) throw new NotEligibleError(reason);
        const discount = discountFor(recompense, amount);
        return { Recompense: recompense, Amount: amount, Discount: discount, Total: amount - discount };
    }
}
