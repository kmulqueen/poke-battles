interface PageHeaderProps {
  title?: string;
}

function PageHeader({ title = "Poke Battles" }: PageHeaderProps) {
  return (
    <header className="pt-8 pb-4 sm:pt-12 sm:pb-6 md:pt-16 md:pb-8 text-center">
      <h1 className="text-4xl sm:text-5xl md:text-6xl font-bold text-white drop-shadow-lg tracking-tight">
        {title}
      </h1>
    </header>
  );
}

export default PageHeader;
