import React from 'react';
import { Tabs } from '#/ui/tabs';
import db from '#/lib/db';
Tabs
export default async function Layout({
  children,
}: {
  children: React.ReactNode;
}) {
  const demo = db.demo.find({ where: { slug: 'route-groups' } });
  const sections = db.section.findMany({ limit: 1 });

  return (
      

      <div>{children}</div>
  );
}