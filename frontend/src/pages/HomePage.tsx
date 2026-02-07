import PageHeader from "../components/layout/PageHeader";
import GradientPageWrapper from "../components/layout/GradientPageWrapper";
import UsernameForm from "../features/auth/components/UsernameForm";
import { useUsernameForm } from "../features/auth/hooks/useUsernameForm";

function HomePage() {
  const { username, setUsername, handleSubmit, isValid } = useUsernameForm();

  return (
    <GradientPageWrapper>
      <PageHeader />
      <main id="main-content" className="flex-1 flex items-center justify-center px-4 py-8 sm:px-6 md:px-8">
        <div className="w-full max-w-sm sm:max-w-md bg-surface rounded-2xl shadow-xl p-6 sm:p-8 md:p-10">
          <p className="text-center text-neutral-600 text-base sm:text-lg mb-6 sm:mb-8">
            Welcome to Poke Battles! Enter a username below.
          </p>
          <section aria-label="Username entry">
            <UsernameForm
              username={username}
              onUsernameChange={setUsername}
              onSubmit={handleSubmit}
              disabled={!isValid}
            />
          </section>
        </div>
      </main>
    </GradientPageWrapper>
  );
}

export default HomePage;
