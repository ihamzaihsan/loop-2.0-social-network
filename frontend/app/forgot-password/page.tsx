import { Suspense } from "react";
import RecoveryForm from "../../components/RecoveryForm";
export default function Page() {
  return (
    <Suspense fallback={<p>Loading?</p>}>
      <RecoveryForm />
    </Suspense>
  );
}
