function getOffset(page: number, limit: number): number {
    if (page <= 0) {
        page = 1;
    }
    return (page - 1) * limit;
}


export { getOffset }