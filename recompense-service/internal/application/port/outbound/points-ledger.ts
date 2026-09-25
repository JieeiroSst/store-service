export interface EarnResult {
    points: number;
    balance: number;
    duplicate: boolean;
}

export interface PointsLedger {
    earn(id: string, memberId: number, key: string, points: number): Promise<EarnResult>;
    balance(memberId: number): Promise<number>;
}
