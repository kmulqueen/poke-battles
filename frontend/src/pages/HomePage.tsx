import PageHeader from "../components/layout/PageHeader";
import UsernameForm from "../features/auth/components/UsernameForm";
import { useUsernameForm } from "../features/auth/hooks/useUsernameForm";

function HomePage() {
  const { username, setUsername, handleSubmit, isValid } = useUsernameForm();

  return (
    <div className="min-h-dvh bg-gradient-to-br from-primary-500 via-primary-600 to-secondary-700 flex flex-col">
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
    </div>
  );
}

export default HomePage;
