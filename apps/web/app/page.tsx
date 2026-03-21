import { SignInButton, UserButton, auth } from "@clerk/nextjs";
import Link from "next/link";

export default async function Home() {
  const { userId } = await auth();

  return (
    <main className="min-h-screen bg-gradient-to-b from-slate-950 to-slate-900">
      <nav className="flex items-center justify-between px-6 py-4 border-b border-slate-800">
        <div className="text-2xl font-bold text-white">🚀 Mission Control</div>
        <div>
          {userId ? (
            <>
              <Link href="/dashboard" className="mr-4 text-slate-300 hover:text-white">
                Dashboard
              </Link>
              <UserButton />
            </>
          ) : (
            <SignInButton />
          )}
        </div>
      </nav>

      <section className="px-6 py-24 text-center max-w-3xl mx-auto">
        <h1 className="text-5xl font-bold text-white mb-4">
          Develop in the Cloud
        </h1>
        <p className="text-xl text-slate-300 mb-8">
          Connect your GitHub repos, spin up isolated VMs, and run Claude Code in your browser.
        </p>

        {!userId ? (
          <SignInButton>
            <button className="px-8 py-3 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
              Get Started →
            </button>
          </SignInButton>
        ) : (
          <Link href="/dashboard">
            <button className="px-8 py-3 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
              Open Dashboard →
            </button>
          </Link>
        )}
      </section>

      <section className="grid grid-cols-3 gap-8 px-6 py-24 max-w-5xl mx-auto">
        <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
          <h3 className="text-lg font-bold text-white mb-2">🔗 Connect Repos</h3>
          <p className="text-slate-400">
            Link your GitHub repositories instantly with OAuth
          </p>
        </div>
        <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
          <h3 className="text-lg font-bold text-white mb-2">⚡ Launch VMs</h3>
          <p className="text-slate-400">
            Spin up isolated workspaces in seconds
          </p>
        </div>
        <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
          <h3 className="text-lg font-bold text-white mb-2">💬 Claude Code</h3>
          <p className="text-slate-400">
            Run AI-powered development in the browser
          </p>
        </div>
      </section>

      <footer className="border-t border-slate-800 px-6 py-6 text-center text-slate-500 text-sm">
        <p>Mission Control — Cloud Platform for Developers</p>
      </footer>
    </main>
  );
}
