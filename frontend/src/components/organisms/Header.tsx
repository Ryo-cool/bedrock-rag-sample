import React from 'react';

const Header: React.FC = () => {
  return (
    <header className="bg-gray-200 text-gray-700 px-4 py-3 shadow-[5px_5px_10px_#bebebe,_-5px_-5px_10px_#ffffff]">
      {/* 上下の影を調整: 5pxずらし, 10pxぼかし, 明るい影(#ffffff), 暗い影(#bebebe) - 色は背景(gray-200)に合わせて調整 */}
      <div className="container mx-auto flex justify-between items-center">
        <h1 className="text-xl font-semibold text-gray-600">Bedrock RAG Sample</h1>
        {/* 必要に応じてナビゲーションリンクなどを追加 */}
        <nav>
          {/* <button className="px-4 py-2 rounded-lg bg-gray-200 text-gray-600 shadow-[inset_3px_3px_5px_#bebebe,inset_-3px_-3px_5px_#ffffff] hover:shadow-[3px_3px_5px_#bebebe,_-3px_-3px_5px_#ffffff]">Link 1</button> */}
          {/* <a href="#" className="px-3 py-2 rounded hover:bg-gray-700">Link 2</a> */}
        </nav>
      </div>
    </header>
  );
};

export default Header; 