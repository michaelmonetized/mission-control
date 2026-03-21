import { SignedIn, SignedOut, SignInButton, UserButton } from "@clerk/nextjs";
import Link from "next/link";

export default function HomePage() {
  return (
    <div className="space-y-12">
      <SignedIn>
        <div className="flex items-center justify-between">
          <h1 className="text-4xl font-bold">Welcome to Mission Control Cloud</h1>
          <UserButton />
        </div>
        <p className="text-lg text-zinc-400">
          Connect your GitHub repos, launch isolated VPS workspaces, and code with Claude.
        </p>
        <Link
          href="/dashboard"
          className="inline-block rounded-lg bg-blue-600 px-6 py-3 font-medium text-white hover:bg-blue-700"
        >
          Go to Dashboard
        </Link>
      </SignedIn>

      <SignedOut>
        <div className="flex flex-col items-center justify-center space-y-8 py-20">
          <div className="text-center">
            <h1 className="mb-4 text-5xl font-bold">🚀 Mission Control Cloud</h1>
            <p className="mb-8 text-xl text-zinc-400">
              Unified project dashboard with isolated VPS workspaces. Code faster with Claude.
            </p>
          </div>

          <div className="space-y-4 text-center">
            <h2 className="text-2xl font-semibold">Features</h2>
            <ul className="space-y-2 text-zinc-300">
              <li>✅ Connect GitHub repositories (public & private)</li>
              <li>✅ Launch isolated VPS workspaces per repo</li>
              <li>✅ Integrated Claude Code IDE in the browser</li>
              <li>✅ Pay-as-you-go compute billing</li>
              <li>✅ Free tier: 100 minutes/month</li>
            </ul>
          </div>

          <SignInButton>
            <button className="rounded-lg bg-blue-600 px-8 py-4 text-lg font-bold text-white hover:bg-blue-700">
              Sign In with GitHub
            </button>
          </SignInButton>
        </div>
      </SignedOut>
    </div>
  );
}
