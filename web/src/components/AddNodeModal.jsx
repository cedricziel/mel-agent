import { useState } from 'react';

function AddNodeModal({ isOpen, categories, onClose, onAddNode }) {
  const [search, setSearch] = useState('');

  if (!isOpen) return null;

  const filteredCategories = categories.filter(({ types }) =>
    types.some(
      (type) =>
        type.label.toLowerCase().includes(search.toLowerCase()) ||
        type.type.toLowerCase().includes(search.toLowerCase())
    )
  );

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-96 max-h-96 overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-bold">Add Node</h2>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700"
          >
            ✕
          </button>
        </div>

        <input
          type="text"
          placeholder="Search nodes..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full border rounded px-3 py-2 mb-4"
        />

        {filteredCategories.map(({ category, types }) => (
          <div key={category} className="mb-4">
            <h3 className="font-semibold text-sm text-gray-600 mb-2">
              {category}
            </h3>
            {types
              .filter(
                (type) =>
                  type.label.toLowerCase().includes(search.toLowerCase()) ||
                  type.type.toLowerCase().includes(search.toLowerCase())
              )
              .map((type) => (
                <button
                  key={type.type}
                  onClick={() => {
                    onAddNode(type.type);
                    onClose();
                  }}
                  className="w-full text-left p-2 hover:bg-gray-100 rounded"
                >
                  <div className="font-medium">{type.label}</div>
                  {type.description && (
                    <div className="text-sm text-gray-500">
                      {type.description}
                    </div>
                  )}
                </button>
              ))}
          </div>
        ))}
      </div>
    </div>
  );
}

export default AddNodeModal;
