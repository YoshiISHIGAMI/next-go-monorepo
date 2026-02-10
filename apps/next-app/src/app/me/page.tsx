import { redirect } from "next/navigation";
import { env } from "@/shared/config/env";
import { auth } from "@/shared/lib/auth";

type User = {
  id: number;
  email: string;
  name: string | null;
  created_at: string;
};

async function getUser(userId: string): Promise<User | null> {
  try {
    const response = await fetch(`${env.apiBaseUrl}/users`, {
      cache: "no-store",
    });

    if (!response.ok) {
      return null;
    }

    const users: User[] = await response.json();
    return users.find((u) => u.id === Number(userId)) ?? null;
  } catch {
    return null;
  }
}

export default async function MePage() {
  const session = await auth();

  if (!session?.user?.id) {
    redirect("/login");
  }

  const user = await getUser(session.user.id);

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <main className="flex flex-col items-center gap-6 p-8">
        <h1 className="text-2xl font-bold text-foreground">マイページ</h1>

        {user ? (
          <div className="flex flex-col gap-4 rounded-lg border p-6">
            <div>
              <p className="text-sm text-muted-foreground">ユーザーID</p>
              <p className="text-foreground">{user.id}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">名前</p>
              <p className="text-foreground">{user.name ?? "未設定"}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">メールアドレス</p>
              <p className="text-foreground">{user.email}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">登録日</p>
              <p className="text-foreground">
                {new Date(user.created_at).toLocaleDateString("ja-JP")}
              </p>
            </div>
          </div>
        ) : (
          <p className="text-destructive">ユーザー情報の取得に失敗しました</p>
        )}
      </main>
    </div>
  );
}
