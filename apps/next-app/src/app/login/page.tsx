import { Button } from "@/components/ui/button";
import { signIn } from "@/shared/lib/auth";

export default function LoginPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="flex flex-col items-center gap-6 p-8">
        <h1 className="text-2xl font-bold text-foreground">ログイン</h1>
        <form
          action={async () => {
            "use server";
            await signIn("github", { redirectTo: "/" });
          }}
        >
          <Button type="submit" size="lg">
            Sign in with GitHub
          </Button>
        </form>
      </div>
    </div>
  );
}
