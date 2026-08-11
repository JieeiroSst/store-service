function hashString(seed: string): number {
  let hash = 0;
  for (let i = 0; i < seed.length; i++) {
    hash = (hash << 5) - hash + seed.charCodeAt(i);
    hash |= 0;
  }
  return Math.abs(hash);
}

const GRADIENTS = [
  ["#8b5cf6", "#ec4899"], // violet -> pink
  ["#6366f1", "#06b6d4"], // indigo -> cyan
  ["#f43f5e", "#f97316"], // rose -> orange
  ["#0ea5e9", "#22c55e"], // sky -> green
  ["#a855f7", "#3b82f6"], // purple -> blue
  ["#ec4899", "#f59e0b"], // pink -> amber
  ["#14b8a6", "#6366f1"], // teal -> indigo
];

export function gradientFor(seed: string): string {
  const [from, to] = GRADIENTS[hashString(seed) % GRADIENTS.length];
  return `linear-gradient(135deg, ${from}, ${to})`;
}

const USERNAME_COLORS = [
  "#f87171",
  "#fb923c",
  "#facc15",
  "#4ade80",
  "#2dd4bf",
  "#38bdf8",
  "#818cf8",
  "#c084fc",
  "#f472b6",
];

export function colorForUsername(seed: string): string {
  return USERNAME_COLORS[hashString(seed) % USERNAME_COLORS.length];
}

export function initialsFor(name: string): string {
  const trimmed = name.trim();
  if (!trimmed) return "?";
  return trimmed.slice(0, 2).toUpperCase();
}
