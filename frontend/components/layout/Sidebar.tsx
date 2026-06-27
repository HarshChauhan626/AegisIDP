"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Server,
  GitBranch,
  FileText,
  Clock,
  Activity,
  ClipboardList,
  Users,
  Zap,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/hooks/useAuth";

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard, permission: "metrics:view" },
  { href: "/environments", label: "Environments", icon: Server, permission: "workflow:view" },
  { href: "/workflows", label: "Workflows", icon: GitBranch, permission: "workflow:view" },
  { href: "/templates", label: "Templates", icon: FileText, permission: "template:manage" },
  { href: "/schedules", label: "Schedules", icon: Clock, permission: "schedule:manage" },
  { href: "/events", label: "Events", icon: Activity, permission: "workflow:view" },
  { href: "/audit", label: "Audit Log", icon: ClipboardList, permission: "audit:view" },
  { href: "/users", label: "Users", icon: Users, permission: "user:manage" },
];

export function Sidebar() {
  const pathname = usePathname();
  const { hasPermission } = useAuth();

  return (
    <aside className="flex flex-col w-64 min-h-screen bg-gray-950 border-r border-gray-800">
      {/* Logo */}
      <div className="flex items-center gap-3 px-6 py-5 border-b border-gray-800">
        <div className="flex items-center justify-center w-9 h-9 rounded-lg bg-gradient-to-br from-violet-600 to-indigo-600">
          <Zap className="w-5 h-5 text-white" />
        </div>
        <div>
          <span className="font-bold text-white text-sm tracking-wide">AegisIDP</span>
          <p className="text-xs text-gray-500">Platform Orchestrator</p>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-3 py-4 space-y-1">
        {navItems.map((item) => {
          if (!hasPermission(item.permission)) return null;
          const Icon = item.icon;
          const isActive = pathname === item.href || pathname.startsWith(item.href + "/");
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-150",
                isActive
                  ? "bg-violet-600/20 text-violet-400 border border-violet-600/30"
                  : "text-gray-400 hover:text-white hover:bg-gray-800"
              )}
            >
              <Icon className="w-4 h-4 flex-shrink-0" />
              {item.label}
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="px-3 py-4 border-t border-gray-800">
        <p className="text-xs text-gray-600 px-3">v1.0.0 — Phase 1</p>
      </div>
    </aside>
  );
}
