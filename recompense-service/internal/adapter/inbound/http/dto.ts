export class RecompenseRequest {
    MemberID: number | undefined;
    Source: string | undefined;
    SourceID: string | undefined;
    Tier: number | undefined;
    PointsPerPurchase: number | undefined;
    MinimumPoints: number | undefined;
    CodeDeReduction: string | undefined;
    ReductionDescription: string | undefined;
    ReductionType?: "fixed" | "percent";
    ReductionValue: number | undefined;
    MaxReduction?: number;
    ExpiryDate: number | undefined;
}
