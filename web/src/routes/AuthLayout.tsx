import { ParentProps, children } from "solid-js";
import { AuthHeader } from "~/components/layout/AuthHeader";

export function AuthLayout(props: ParentProps) {
  const safeChildren = children(() => props.children);

  return (
    <div>
      <AuthHeader />
      <main>
        {safeChildren()}
      </main>
    </div>
  );
}
