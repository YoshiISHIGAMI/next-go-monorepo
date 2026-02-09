import Link from "next/link";
import { Button } from "@/components/ui/button";
import { auth, signOut } from "@/shared/lib/auth";

export default async function Home() {
  const session = await auth();

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <main className="flex flex-col items-center gap-8 p-8">
        <h1 className="text-3xl font-bold text-foreground">
          Next.js + shadcn/ui
        </h1>

        {session ? (
          <div className="flex flex-col items-center gap-4">
            <p className="text-foreground">
              ログイン中: {session.user?.name ?? session.user?.email}
            </p>
            <form
              action={async () => {
                "use server";
                await signOut({ redirectTo: "/" });
              }}
            >
              <Button type="submit" variant="outline">
                Sign out
              </Button>
            </form>
          </div>
        ) : (
          <Button asChild>
            <Link href="/login">Sign in with GitHub</Link>
          </Button>
        )}
      </main>
    </div>
  );
}
