import { Route, Router } from "@solidjs/router";
import { WaitlistRoute } from "./Waitlist";
import { AuthLayout } from "./AuthLayout";

export function RootLayout() {
  return (
      <Router>
        <Route path="/" component={WaitlistRoute} />
        <Route path="/auth" component={AuthLayout}>
          <Route path="/signup" component={() => null} />
          <Route path="/login" component={() => null} />
          <Route path="/forgot-password" component={() => null} />
          <Route path="/reset-password" component={() => null} />
        </Route>
      </Router>
  );
}
