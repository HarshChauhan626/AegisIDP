"use client";

import { useState } from "react";
import Editor from "@monaco-editor/react";
import { FileCode2, Save } from "lucide-react";

const DEFAULT_TEMPLATE = `name: provision-environment
description: "Creates a full development environment"
steps:
  - name: create_namespace
    executor: namespace
  - name: provision_db
    executor: postgres
    depends_on: [create_namespace]
  - name: provision_cache
    executor: redis
    depends_on: [create_namespace]
`;

export default function TemplatesPage() {
  const [yamlContent, setYamlContent] = useState(DEFAULT_TEMPLATE);
  const [activeTemplate, setActiveTemplate] = useState("provision-environment");

  return (
    <div className="flex flex-col h-[calc(100vh-8rem)]">
      {/* Header */}
      <div className="flex items-center justify-between mb-6 shrink-0">
        <div>
          <h2 className="text-2xl font-bold text-white">Workflow Templates</h2>
          <p className="text-gray-400 text-sm mt-1">Author and manage YAML workflow definitions</p>
        </div>
        <button
          className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white text-sm font-medium rounded-lg transition-all shadow-lg shadow-violet-500/20"
        >
          <Save className="w-4 h-4" />
          Save Template
        </button>
      </div>

      {/* Editor Layout */}
      <div className="flex flex-1 gap-6 min-h-0">
        {/* Sidebar */}
        <div className="w-64 shrink-0 flex flex-col gap-2">
          <div className="text-sm font-medium text-gray-400 mb-2">Saved Templates</div>
          <button
            className="flex items-center gap-3 px-3 py-2 bg-indigo-600/10 text-indigo-400 border border-indigo-500/20 rounded-lg text-sm text-left transition-colors"
          >
            <FileCode2 className="w-4 h-4 shrink-0" />
            <span className="truncate">provision-environment</span>
          </button>
          <button
            className="flex items-center gap-3 px-3 py-2 text-gray-400 hover:bg-gray-800/50 hover:text-gray-300 rounded-lg text-sm text-left transition-colors"
          >
            <FileCode2 className="w-4 h-4 shrink-0" />
            <span className="truncate">delete-environment</span>
          </button>
        </div>

        {/* Monaco Editor Container */}
        <div className="flex-1 bg-[#1e1e1e] rounded-xl border border-gray-800 overflow-hidden relative shadow-inner">
          <Editor
            height="100%"
            defaultLanguage="yaml"
            theme="vs-dark"
            value={yamlContent}
            onChange={(val) => setYamlContent(val || "")}
            options={{
              minimap: { enabled: false },
              fontSize: 14,
              lineHeight: 1.5,
              padding: { top: 16, bottom: 16 },
              scrollBeyondLastLine: false,
              fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
            }}
            loading={
              <div className="absolute inset-0 flex items-center justify-center bg-[#1e1e1e] text-gray-400 text-sm">
                Loading editor...
              </div>
            }
          />
        </div>
      </div>
    </div>
  );
}
