import { Button } from "@/components/ui/button";

export default function Home() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <main className="flex flex-col items-center gap-8 p-8">
        <h1 className="text-3xl font-bold text-foreground">
          Next.js + shadcn/ui
        </h1>
        <div className="flex flex-wrap gap-4">
          <Button>Default</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="outline">Outline</Button>
          <Button variant="destructive">Destructive</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="link">Link</Button>
        </div>
      </main>
    </div>
  );
}
