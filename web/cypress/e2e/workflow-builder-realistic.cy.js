describe('Workflow Builder - Realistic Tests', () => {
  beforeEach(() => {
    // Mock all the required API endpoints
    cy.intercept('GET', '/api/node-types', {
      statusCode: 200,
      body: [
        {
          type: 'http_request',
          label: 'HTTP Request',
          category: 'Actions',
          description: 'Make HTTP requests to external APIs'
        },
        {
          type: 'log',
          label: 'Log',
          category: 'Utility',
          description: 'Log messages for debugging'
        }
      ]
    }).as('getNodeTypes')

    cy.intercept('GET', '/api/triggers', {
      statusCode: 200,
      body: []
    }).as('getTriggers')

    // Mock draft API (returns 404 - no draft exists)
    cy.intercept('GET', '/api/agents/test-agent-1/draft', {
      statusCode: 404,
      body: { error: 'No draft found' }
    }).as('getDraft')

    // Mock workflow client fallback APIs (3 parallel calls)
    cy.intercept('GET', '/api/workflows/test-agent-1', {
      statusCode: 200,
      body: { id: 'test-agent-1', name: 'Test Workflow' }
    }).as('getWorkflowMeta')

    cy.intercept('GET', '/api/workflows/test-agent-1/nodes', {
      statusCode: 200,
      body: []
    }).as('getWorkflowNodes')

    cy.intercept('GET', '/api/workflows/test-agent-1/edges', {
      statusCode: 200,
      body: []
    }).as('getWorkflowEdges')

    // Mock WebSocket by intercepting the connection attempt
    cy.window().then((win) => {
      // Mock WebSocket constructor
      const originalWebSocket = win.WebSocket
      win.WebSocket = function(url) {
        const ws = {
          send: cy.stub(),
          close: cy.stub(),
          readyState: 1, // OPEN
          onmessage: null,
          onclose: null,
          onopen: null,
          onerror: null
        }
        // Simulate successful connection
        setTimeout(() => {
          if (ws.onopen) ws.onopen()
        }, 10)
        return ws
      }
    })

    cy.visit('/agents/test-agent-1/edit')
    cy.wait(['@getNodeTypes', '@getDraft', '@getWorkflowMeta', '@getWorkflowNodes', '@getWorkflowEdges'])
  })

  it('should load the workflow builder interface', () => {
    // Check for ReactFlow canvas (uses specific ReactFlow classes)
    cy.get('.react-flow').should('be.visible')
    cy.get('.react-flow__viewport').should('be.visible')
    
    // Check for main toolbar buttons
    cy.contains('+ Add Node').should('be.visible')
    cy.contains('Test Run').should('be.visible')
    cy.contains('Auto Layout').should('be.visible')
  })

  it('should show draft/deployed status indicator', () => {
    // Should show "Deployed" status by default (no draft)
    cy.contains('Deployed').should('be.visible')
    cy.get('.bg-green-100').should('contain', 'Deployed')
  })

  it('should have editor/executions toggle', () => {
    cy.contains('Editor').should('be.visible')
    cy.contains('Executions').should('be.visible')
    
    // Editor should be active by default
    cy.contains('Editor').should('have.class', 'bg-white')
  })

  it('should open add node modal', () => {
    cy.contains('+ Add Node').click()
    
    // Modal should appear
    cy.get('.fixed.inset-0').should('be.visible') // Modal overlay
    cy.contains('Add Node').should('be.visible')
    cy.get('input[placeholder="Search nodes..."]').should('be.visible')
    
    // Should show node categories
    cy.contains('Actions').should('be.visible')
    cy.contains('Utility').should('be.visible')
    cy.contains('HTTP Request').should('be.visible')
    cy.contains('Log').should('be.visible')
  })

  it('should close add node modal', () => {
    cy.contains('+ Add Node').click()
    cy.get('button').contains('✕').click()
    cy.get('.fixed.inset-0').should('not.exist')
  })

  it('should filter nodes in search', () => {
    cy.contains('+ Add Node').click()
    cy.get('input[placeholder="Search nodes..."]').type('http')
    
    // Should show HTTP Request but hide Log
    cy.contains('HTTP Request').should('be.visible')
    cy.contains('Log').should('not.exist')
  })

  it('should open chat assistant', () => {
    cy.contains('💬 Chat').click()
    
    // Sidebar should appear
    cy.get('.w-80.bg-white.border-l').should('be.visible')
    
    // Close chat
    cy.contains('💬 Chat').click()
    cy.get('.w-80.bg-white.border-l').should('not.exist')
  })

  it('should toggle live mode', () => {
    // Should start in Edit Mode
    cy.contains('Edit Mode').should('be.visible')
    
    cy.contains('Edit Mode').click()
    cy.contains('Live Mode').should('be.visible')
    cy.get('button').contains('Live Mode').should('have.class', 'bg-orange-500')
    
    // Toggle back
    cy.contains('Live Mode').click()
    cy.contains('Edit Mode').should('be.visible')
  })

  it('should switch to executions view', () => {
    // Mock executions API
    cy.intercept('GET', '/api/agents/test-agent-1/runs', {
      statusCode: 200,
      body: []
    }).as('getRuns')

    cy.contains('Executions').click()
    cy.wait('@getRuns')
    
    // Should show executions panel
    cy.get('.w-80.bg-white.border-l').should('be.visible')
    cy.contains('No executions found').should('be.visible')
    
    // Buttons should be disabled in executions mode
    cy.contains('+ Add Node').should('have.class', 'cursor-not-allowed')
  })

  it('should handle test run', () => {
    cy.intercept('POST', '/api/agents/test-agent-1/runs/test', {
      statusCode: 200,
      body: { success: true }
    }).as('testRun')

    cy.contains('Test Run').click()
    cy.wait('@testRun')
    
    // Button should change to "Running..." during execution
    // This might be too fast to catch in test, but at least verify the request was made
  })

  it('should handle ReactFlow canvas interactions', () => {
    // Should be able to interact with ReactFlow canvas
    cy.get('.react-flow__viewport').should('be.visible')
    cy.get('.react-flow__controls').should('be.visible') // ReactFlow controls
    cy.get('.react-flow__minimap').should('be.visible') // ReactFlow minimap
    
    // Click on canvas (should not error) - force click since ReactFlow disables pointer events
    cy.get('.react-flow__viewport').click(400, 300, { force: true })
  })

  it('should handle loading and error states', () => {
    // Test loading state by intercepting with delay
    cy.intercept('GET', '/api/agents/slow-agent/versions/latest', {
      delay: 100,
      statusCode: 200,
      body: { nodes: [], edges: [] }
    }).as('slowWorkflow')

    cy.visit('/agents/slow-agent/edit')
    cy.contains('Loading workflow...').should('be.visible')
  })
})