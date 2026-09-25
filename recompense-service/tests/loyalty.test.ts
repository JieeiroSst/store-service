import { describe, expect, test } from "bun:test";

import { discountFor, earnPoints, memberStatus, selectBest, tierFor } from "../internal/domain/loyalty";
import type { Recompense } from "../internal/domain/recompense";

const r = (o: Partial<Recompense>): Recompense => ({ ID: "1", MemberID: 1, ...o }) as Recompense;

describe("tiers", () => {
    test("boundaries", () => {
        expect(tierFor(0).tier).toBe(1);
        expect(tierFor(999).tier).toBe(1);
        expect(tierFor(1000).tier).toBe(2);
        expect(tierFor(20000).tier).toBe(4);
    });
    test("progress to next tier", () => {
        expect(memberStatus(1, 900)).toMatchObject({ Tier: 1, NextTier: 2, PointsToNextTier: 100 });
        expect(memberStatus(1, 25000).NextTier).toBeUndefined();
    });
});

describe("earnPoints", () => {
    test("multiplier by tier", () => {
        expect(earnPoints(1000, 1)).toBe(100);
        expect(earnPoints(1000, 2)).toBe(125);
        expect(earnPoints(1000, 4)).toBe(200);
    });
    test("invalid amounts earn nothing", () => {
        expect(earnPoints(-5, 1)).toBe(0);
        expect(earnPoints(NaN, 1)).toBe(0);
    });
});

describe("discountFor", () => {
    test("fixed is capped by the purchase", () => {
        expect(discountFor(r({ ReductionValue: 50 }), 30)).toBe(30);
    });
    test("percent honors MaxReduction", () => {
        expect(discountFor(r({ ReductionType: "percent", ReductionValue: 20, MaxReduction: 15 }), 200)).toBe(15);
        expect(discountFor(r({ ReductionType: "percent", ReductionValue: 20 }), 200)).toBe(40);
    });
});

describe("selectBest", () => {
    const status = memberStatus(1, 1500); // silver
    const now = 1_000;
    test("skips expired, too-high tier and low points", () => {
        const list = [
            r({ ID: "1", ReductionValue: 90, ExpiryDate: 500 }),
            r({ ID: "2", ReductionValue: 80, Tier: 3 }),
            r({ ID: "3", ReductionValue: 70, MinimumPoints: 2000 }),
            r({ ID: "4", ReductionValue: 10 }),
        ];
        expect(selectBest(list, status, 100, now)?.recompense.ID).toBe("4");
    });
    test("prefers larger discount, then earlier expiry, then lower id", () => {
        expect(selectBest([r({ ID: "1", ReductionValue: 5 }), r({ ID: "2", ReductionValue: 9 })], status, 100, now)?.recompense.ID).toBe("2");
        expect(selectBest([r({ ID: "1", ReductionValue: 5 }), r({ ID: "2", ReductionValue: 5, ExpiryDate: 5000 })], status, 100, now)?.recompense.ID).toBe("2");
        expect(selectBest([r({ ID: "9", ReductionValue: 5 }), r({ ID: "3", ReductionValue: 5 })], status, 100, now)?.recompense.ID).toBe("3");
    });
    test("nothing usable", () => {
        expect(selectBest([], status, 100, now)).toBeUndefined();
    });
});
