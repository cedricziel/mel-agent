import { useState, useEffect, useRef } from 'react';
import { assistantApi, nodeTypesApi } from '../api/client';
// Temporarily disabled due to React 19 compatibility issues
// import ReactMarkdown from 'react-markdown';
// import remarkGfm from 'remark-gfm';

// ChatAssistant provides a modal chat interface for users to interact with an AI assistant
// and supports function-based tools to modify the workflow graph.
export default function ChatAssistant({
  agentId,
  onAddNode,
  onConnectNodes,
  onGetWorkflow,
  onClose,
  inline = false,
}) {
  const [messages, setMessages] = useState([
    {
      role: 'system',
      content:
        'You are a helpful AI assistant for the workflow builder. You can call functions `list_node_types`, `get_node_type_schema`, and `get_workflow` to inspect the current graph, and use `add_node` and `connect_nodes` to modify the workflow graph.',
    },
  ]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const scrollRef = useRef();
  const abortControllerRef = useRef(null);

  // Maximum message history to prevent unbounded growth
  const MAX_MESSAGES = 100;

  // Auto-scroll on new messages
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  // Limit message history to prevent memory bloat
  useEffect(() => {
    if (messages.length > MAX_MESSAGES) {
      setMessages((prev) => [prev[0], ...prev.slice(-MAX_MESSAGES + 1)]);
    }
  }, [messages]);

  const sendMessage = async () => {
    const text = input.trim();
    if (!text) return;

    // Abort any existing request
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // Create new abort controller for this conversation
    abortControllerRef.current = new AbortController();

    // Prepare a mutable copy of the conversation
    setIsLoading(true);
    let convo = [...messages, { role: 'user', content: text }];
    setMessages(convo);
    setInput('');

    try {
      // Loop until the model returns a non-function message
      // Add safety limit to prevent infinite loops
      let loopCount = 0;
      const MAX_FUNCTION_CALLS = 10;

      while (loopCount < MAX_FUNCTION_CALLS) {
        loopCount++;
        const res = await assistantApi.assistantChat({
          messages: convo,
          workflow_id: agentId, // Optional workflow context
        });
        const msg = res.data.choices[0].message;
        if (msg.function_call) {
          // Append the function_call message
          setMessages((ms) => [
            ...ms,
            {
              role: 'assistant',
              content: '',
              function_call: msg.function_call,
            },
          ]);
          convo = [...convo, msg];
          // Execute the tool
          const fnName = msg.function_call.name;
          const fnArgs = JSON.parse(msg.function_call.arguments || '{}');
          let result;
          if (fnName === 'add_node') result = onAddNode(fnArgs);
          else if (fnName === 'connect_nodes') result = onConnectNodes(fnArgs);
          else if (fnName === 'list_node_types') {
            const r = await nodeTypesApi.listNodeTypes();
            result = r.data;
          } else if (fnName === 'get_node_type_schema') {
            const { type } = fnArgs;
            const r = await nodeTypesApi.getNodeTypeSchema(type);
            result = r.data;
          } else if (fnName === 'get_node_definition') {
            const { type } = fnArgs;
            const r = await nodeTypesApi.listNodeTypes();
            result = (r.data || []).find((nt) => nt.type === type) || {};
          } else if (fnName === 'get_workflow') result = onGetWorkflow();
          const resultContent = JSON.stringify(result || {});
          // Append the function result
          setMessages((ms) => [
            ...ms,
            { role: 'function', name: fnName, content: resultContent },
          ]);
          convo = [
            ...convo,
            { role: 'function', name: fnName, content: resultContent },
          ];
          // Continue loop for next call
        } else {
          // Final assistant response
          setMessages((ms) => [
            ...ms,
            { role: 'assistant', content: msg.content },
          ]);
          break;
        }
      }

      // If we exit the loop due to max function calls, add a warning
      if (loopCount >= MAX_FUNCTION_CALLS) {
        setMessages((ms) => [
          ...ms,
          {
            role: 'assistant',
            content:
              'Warning: Maximum function call limit reached. Please try a simpler request.',
          },
        ]);
      }
    } catch (err) {
      console.error(err);
      setMessages((ms) => [
        ...ms,
        { role: 'assistant', content: 'Error: Unable to contact assistant.' },
      ]);
    } finally {
      setIsLoading(false);
    }
  };

  // Summarize function results for UI display
  const summarizeResult = (fnName, result) => {
    try {
      switch (fnName) {
        case 'list_node_types':
          if (Array.isArray(result)) {
            const list = result.slice(0, 5).map((nt) => `\`${nt.type}\``);
            const more =
              result.length > 5 ? ` and ${result.length - 5} more` : '';
            return `**list_node_types** returned ${result.length} types: ${list.join(', ')}${more}.`;
          }
          break;
        case 'get_node_type_schema':
          if (result.properties && typeof result.properties === 'object') {
            const props = Object.keys(result.properties).map((p) => `\`${p}\``);
            return `Schema has properties: ${props.join(', ')}.`;
          }
          break;
        case 'get_workflow':
          if (result.nodes && result.edges) {
            return `Current workflow has ${result.nodes.length} nodes and ${result.edges.length} edges.`;
          }
          break;
      }
    } catch {
      // fallback
    }
    const text = JSON.stringify(result, null, 2);
    return text.length > 300 ? text.slice(0, 300) + '...' : text;
  };

  if (inline) {
    // Inline mode: rendered inside parent sidebar without its own header
    return (
      <div className="flex flex-col h-full p-2 overflow-auto">
        <div ref={scrollRef} className="flex-1 overflow-auto space-y-2 text-sm">
          {messages.map((msg, idx) => {
            let display;
            if (msg.role === 'function') {
              try {
                const data = JSON.parse(msg.content);
                display = summarizeResult(msg.name, data);
              } catch {
                display = msg.content;
              }
            } else {
              display = msg.content;
            }
            return (
              <div
                key={idx}
                className={
                  msg.role === 'user'
                    ? 'text-right'
                    : msg.role === 'assistant'
                      ? 'text-left text-gray-700'
                      : 'text-left text-blue-500'
                }
              >
                {msg.role === 'assistant' && msg.function_call && (
                  <div className="italic text-gray-500">
                    Function call: {msg.function_call.name}(…)
                  </div>
                )}
                {/* Temporarily use plain text due to React 19 compatibility */}
                <div className="whitespace-pre-wrap">{display}</div>
              </div>
            );
          })}
        </div>
        <div className="p-2 border-t flex">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') sendMessage();
            }}
            placeholder="Type a message..."
            className="flex-1 border rounded px-2 py-1 mr-2"
            disabled={isLoading}
          />
          <button
            onClick={sendMessage}
            disabled={isLoading}
            className="px-3 py-1 bg-indigo-600 text-white rounded disabled:opacity-50"
          >
            Send
          </button>
        </div>
      </div>
    );
  }
  // modal overlay
  return (
    <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center">
      <div className="bg-white rounded shadow-lg w-96 h-3/4 flex flex-col">
        <div className="flex justify-between items-center p-2 border-b">
          <h2 className="text-lg font-bold">AI Assistant</h2>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-800"
          >
            &times;
          </button>
        </div>
        <div
          ref={scrollRef}
          className="p-2 flex-1 overflow-auto space-y-2 text-sm"
        >
          {messages.map((msg, idx) => {
            let display;
            if (msg.role === 'function') {
              try {
                const data = JSON.parse(msg.content);
                display = summarizeResult(msg.name, data);
              } catch {
                display = msg.content;
              }
            } else {
              display = msg.content;
            }
            return (
              <div
                key={idx}
                className={
                  msg.role === 'user'
                    ? 'text-right'
                    : msg.role === 'assistant'
                      ? 'text-left text-gray-700'
                      : 'text-left text-blue-500'
                }
              >
                {msg.role === 'assistant' && msg.function_call && (
                  <div className="italic text-gray-500">
                    Function call: {msg.function_call.name}(…)
                  </div>
                )}
                {/* Temporarily use plain text due to React 19 compatibility */}
                <div className="whitespace-pre-wrap">{display}</div>
              </div>
            );
          })}
        </div>
        <div className="p-2 border-t flex">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') sendMessage();
            }}
            placeholder="Type a message..."
            className="flex-1 border rounded px-2 py-1 mr-2"
            disabled={isLoading}
          />
          <button
            onClick={sendMessage}
            disabled={isLoading}
            className="px-3 py-1 bg-indigo-600 text-white rounded disabled:opacity-50"
          >
            Send
          </button>
        </div>
      </div>
    </div>
  );
}
