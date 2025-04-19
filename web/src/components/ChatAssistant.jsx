import React, { useState, useEffect, useRef } from 'react';
import axios from 'axios';

// ChatAssistant provides a modal chat interface for users to interact with an AI assistant
// and supports function-based tools to modify the workflow graph.
export default function ChatAssistant({ agentId, onAddNode, onConnectNodes, onClose }) {
  const [messages, setMessages] = useState([
    { role: 'system', content: 'You are a helpful AI assistant for the workflow builder. You can call functions `list_node_types` and `get_node_type_schema` to discover available node types and their parameter schemas, and use `add_node` and `connect_nodes` to modify the workflow graph.' }
  ]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const scrollRef = useRef();

  // Auto-scroll on new messages
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const sendMessage = async () => {
    const text = input.trim();
    if (!text) return;
    const userMsg = { role: 'user', content: text };
    const newMessages = [...messages, userMsg];
    setMessages(newMessages);
    setInput('');
    setIsLoading(true);
    try {
      // Initial chat completion, may request a function call
      const res1 = await axios.post(`/api/agents/${agentId}/assistant/chat`, { messages: newMessages });
      const choice1 = res1.data.choices[0].message;
      if (choice1.function_call) {
        // Record the function call
        setMessages(ms => [...ms, { role: 'assistant', content: '', function_call: choice1.function_call }]);
        // Execute the tool locally or fetch definitions
        const fnName = choice1.function_call.name;
        const fnArgs = JSON.parse(choice1.function_call.arguments || '{}');
        let result;
        if (fnName === 'add_node') {
          result = onAddNode(fnArgs);
        } else if (fnName === 'connect_nodes') {
          result = onConnectNodes(fnArgs);
        } else if (fnName === 'list_node_types') {
          // Fetch available node types from backend
          const res = await axios.get(`/api/node-types`);
          result = res.data;
        } else if (fnName === 'get_node_type_schema') {
          // Fetch JSON Schema for a specific node type
          const { type } = fnArgs;
          const res = await axios.get(`/api/node-types/schema/${type}`);
          result = res.data;
        }
        const fnResultContent = JSON.stringify(result || {});
        setMessages(ms => [...ms, { role: 'function', name: fnName, content: fnResultContent }]);
        // Second chat completion to get assistant response
        const followup = await axios.post(`/api/agents/${agentId}/assistant/chat`, {
          messages: [...newMessages, choice1, { role: 'function', name: fnName, content: fnResultContent }]
        });
        const choice2 = followup.data.choices[0].message;
        setMessages(ms => [...ms, { role: 'assistant', content: choice2.content }]);
      } else {
        // Direct content response
        setMessages(ms => [...ms, { role: 'assistant', content: choice1.content }]);
      }
    } catch (err) {
      console.error(err);
      setMessages(ms => [...ms, { role: 'assistant', content: 'Error: Unable to contact assistant.' }]);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center">
      <div className="bg-white rounded shadow-lg w-96 h-3/4 flex flex-col">
        <div className="flex justify-between items-center p-2 border-b">
          <h2 className="text-lg font-bold">AI Assistant</h2>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-800">&times;</button>
        </div>
        <div ref={scrollRef} className="p-2 flex-1 overflow-auto space-y-2 text-sm">
          {messages.map((msg, idx) => (
            <div key={idx} className={msg.role === 'user' ? 'text-right' : msg.role === 'assistant' ? 'text-left text-gray-700' : 'text-left text-blue-500'}>
              {msg.role === 'assistant' && msg.function_call && (
                <div className="italic text-gray-500">Function call: {msg.function_call.name}(…)</div>
              )}
              {msg.content}
            </div>
          ))}
        </div>
        <div className="p-2 border-t flex">
          <input
            type="text"
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') sendMessage(); }}
            placeholder="Type a message..."
            className="flex-1 border rounded px-2 py-1 mr-2"
            disabled={isLoading}
          />
          <button onClick={sendMessage} disabled={isLoading} className="px-3 py-1 bg-indigo-600 text-white rounded disabled:opacity-50">
            Send
          </button>
        </div>
      </div>
    </div>
  );
}