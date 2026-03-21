"use client";

interface Usage {
  minutes: number;
  cost: number;
  remaining: number;
  period: string;
}

export default function UsageDisplay({ usage }: { usage: Usage }) {
  const percentUsed = Math.round((100 - Math.max(0, usage.remaining)) / 1) || 0;
  const isFree = usage.cost === 0;

  return (
    <div className="rounded-lg border border-zinc-700 bg-zinc-900 p-6">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-xl font-semibold">Usage ({usage.period})</h2>
        <span className="text-sm text-zinc-400">{isFree ? "Free Tier" : "Pay-as-you-go"}</span>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        {/* Minutes Used */}
        <div className="space-y-2">
          <div className="text-sm text-zinc-400">Minutes Used</div>
          <div className="text-3xl font-bold">{usage.minutes}</div>
          <div className="text-sm text-zinc-500">of 100 free</div>
          <div className="mt-2 h-2 w-full rounded-full bg-zinc-700">
            <div
              className={`h-full rounded-full transition-all ${
                percentUsed > 90 ? "bg-red-500" : percentUsed > 75 ? "bg-yellow-500" : "bg-green-500"
              }`}
              style={{ width: `${Math.min(percentUsed, 100)}%` }}
            />
          </div>
        </div>

        {/* Cost */}
        <div className="space-y-2">
          <div className="text-sm text-zinc-400">Current Cost</div>
          <div className="text-3xl font-bold">${usage.cost.toFixed(2)}</div>
          <div className="text-sm text-zinc-500">This month</div>
        </div>

        {/* Remaining */}
        <div className="space-y-2">
          <div className="text-sm text-zinc-400">Free Minutes Left</div>
          <div className={`text-3xl font-bold ${usage.remaining > 10 ? "text-green-400" : "text-red-400"}`}>
            {Math.max(0, usage.remaining)}
          </div>
          <div className="text-sm text-zinc-500">Before charges apply</div>
        </div>
      </div>

      {usage.remaining <= 0 && (
        <div className="mt-4 rounded-lg border border-red-700/50 bg-red-900/20 p-3">
          <p className="text-sm text-red-200">
            You've reached your free tier limit. Additional compute will be billed to your Stripe account.
          </p>
        </div>
      )}
    </div>
  );
}
