export type ReductionType = "fixed" | "percent";

export class Recompense {
    ID?: string;
    MemberID: number | undefined;
    Source: string | undefined;
    SourceID: string | undefined;
    Tier: number | undefined;
    PointsPerPurchase: number | undefined;
    MinimumPoints: number | undefined;
    CodeDeReduction: string | undefined;
    ReductionDescription: string | undefined;
    ReductionType?: ReductionType;
    ReductionValue: number | undefined;
    MaxReduction?: number;
    ExpiryDate: number | undefined;
}

export class ValidationError extends Error { }

export function validate(r: Recompense): void {
    if (!Number.isInteger(r.MemberID) || (r.MemberID as number) <= 0) {
        throw new ValidationError("MemberID must be a positive integer");
    }
    if (r.ReductionType !== undefined && r.ReductionType !== "fixed" && r.ReductionType !== "percent") {
        throw new ValidationError("ReductionType must be fixed or percent");
    }
    if (r.ReductionValue !== undefined && (!Number.isFinite(r.ReductionValue) || r.ReductionValue < 0)) {
        throw new ValidationError("ReductionValue must be >= 0");
    }
    if (r.ReductionType === "percent" && (r.ReductionValue ?? 0) > 100) {
        throw new ValidationError("percent ReductionValue must be <= 100");
    }
    if (r.MaxReduction !== undefined && (!Number.isFinite(r.MaxReduction) || r.MaxReduction < 0)) {
        throw new ValidationError("MaxReduction must be >= 0");
    }
}

export class NotEligibleError extends Error { }
