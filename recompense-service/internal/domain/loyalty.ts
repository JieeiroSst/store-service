import type { Recompense } from "./recompense";

export interface TierSpec {
    tier: number;
    name: string;
    minPoints: number;
    multiplier: number;
}

export const tiers: readonly TierSpec[] = [
    { tier: 1, name: "bronze", minPoints: 0, multiplier: 1 },
    { tier: 2, name: "silver", minPoints: 1_000, multiplier: 1.25 },
    { tier: 3, name: "gold", minPoints: 5_000, multiplier: 1.5 },
    { tier: 4, name: "platinum", minPoints: 20_000, multiplier: 2 },
];

export const unitsPerPoint = 10;

export interface MemberStatus {
    MemberID: number;
    Points: number;
    Tier: number;
    TierName: string;
    NextTier?: number;
    PointsToNextTier?: number;
}

export function tierFor(points: number): TierSpec {
    let current = tiers[0]!;
    for (const t of tiers) {
        if (points >= t.minPoints) current = t;
    }
    return current;
}

export function memberStatus(memberId: number, points: number): MemberStatus {
    const current = tierFor(points);
    const next = tiers.find((t) => t.minPoints > points);
    return {
        MemberID: memberId,
        Points: points,
        Tier: current.tier,
        TierName: current.name,
        NextTier: next?.tier,
        PointsToNextTier: next ? next.minPoints - points : undefined,
    };
}

export function earnPoints(amount: number, tier: number): number {
    if (!Number.isFinite(amount) || amount <= 0) return 0;
    const multiplier = tiers.find((t) => t.tier === tier)?.multiplier ?? 1;
    return Math.floor(Math.floor(amount / unitsPerPoint) * multiplier);
}

export function ineligibleReason(r: Recompense, status: MemberStatus, now: number): string | undefined {
    if (r.ExpiryDate !== undefined && r.ExpiryDate <= now) return "expired";
    if (r.Tier !== undefined && status.Tier < r.Tier) return "tier too low";
    if (r.MinimumPoints !== undefined && status.Points < r.MinimumPoints) return "not enough points";
    return undefined;
}

export function discountFor(r: Recompense, amount: number): number {
    if (!Number.isFinite(amount) || amount <= 0) return 0;
    const value = r.ReductionValue ?? 0;
    let discount = r.ReductionType === "percent" ? (amount * value) / 100 : value;
    if (r.ReductionType === "percent" && r.MaxReduction !== undefined) {
        discount = Math.min(discount, r.MaxReduction);
    }
    return Math.round(Math.min(discount, amount) * 100) / 100;
}

export function selectBest(
    candidates: Recompense[],
    status: MemberStatus,
    amount: number,
    now: number,
): { recompense: Recompense; discount: number } | undefined {
    let best: { recompense: Recompense; discount: number } | undefined;
    for (const r of candidates) {
        if (ineligibleReason(r, status, now) !== undefined) continue;
        const discount = discountFor(r, amount);
        if (discount <= 0) continue;
        if (!best || beats(r, discount, best.recompense, best.discount)) best = { recompense: r, discount };
    }
    return best;
}

function beats(r: Recompense, d: number, cur: Recompense, curD: number): boolean {
    if (d !== curD) return d > curD;
    const e = r.ExpiryDate ?? Infinity;
    const curE = cur.ExpiryDate ?? Infinity;
    if (e !== curE) return e < curE;
    return BigInt(r.ID ?? 0) < BigInt(cur.ID ?? 0);
}
