import React from 'react';
import Link from 'next/link';

const Sidebar: React.FC = () => {
  // ニューモフィズム用の影スタイル (背景色 gray-200 に合わせて調整)
  const neumorphismShadow = "shadow-[5px_5px_10px_#bebebe,_-5px_-5px_10px_#ffffff]";
  // const neumorphismInsetShadow = "shadow-[inset_5px_5px_10px_#bebebe,inset_-5px_-5px_10px_#ffffff]";
  const hoverNeumorphismShadow = "hover:shadow-[inset_3px_3px_5px_#bebebe,inset_-3px_-3px_5px_#ffffff]"; // ホバー時は少し凹む効果

  return (
    <aside className={`w-64 bg-gray-200 p-4 h-screen ${neumorphismShadow}`}> {/* 背景色と影 */}
      <nav className="mt-4">
        <ul>
          <li className="mb-3">
            <Link href="/"
                  className={`block px-3 py-2 rounded-lg text-gray-700 ${hoverNeumorphismShadow} transition-shadow duration-200 ease-in-out`}>
              ホーム
            </Link>
          </li>
          <li className="mb-3">
            <Link href="/summarize/text"
                  className={`block px-3 py-2 rounded-lg text-gray-700 ${hoverNeumorphismShadow} transition-shadow duration-200 ease-in-out`}>
              テキスト要約
            </Link>
          </li>
          <li className="mb-3">
            <Link href="/qa"
                  className={`block px-3 py-2 rounded-lg text-gray-700 ${hoverNeumorphismShadow} transition-shadow duration-200 ease-in-out`}>
              Q&A
            </Link>
          </li>
          <li className="mb-3">
            <Link href="/upload"
                  className={`block px-3 py-2 rounded-lg text-gray-700 ${hoverNeumorphismShadow} transition-shadow duration-200 ease-in-out`}>
              ファイルアップロード
            </Link>
          </li>
          {/* 必要に応じて項目を追加 */}
        </ul>
      </nav>
    </aside>
  );
};

export default Sidebar; 