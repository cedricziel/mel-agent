// Custom Cypress commands for MEL Agent testing

// Mock API responses for testing
Cypress.Commands.add('mockAPIEndpoints', () => {
  // Mock agents/workflows endpoint
  cy.intercept('GET', '/api/agents', {
    statusCode: 200,
    body: [
      {
        id: 'test-agent-1',
        name: 'Test Agent',
        description: 'A test workflow',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z'
      }
    ]
  }).as('getAgents')

  // Mock node types endpoint
  cy.intercept('GET', '/api/node-types', {
    statusCode: 200,
    body: [
      {
        type: 'http_request',
        label: 'HTTP Request',
        category: 'Actions',
        parameters: [
          { name: 'url', label: 'URL', type: 'string', required: true },
          { name: 'method', label: 'Method', type: 'select', options: ['GET', 'POST', 'PUT', 'DELETE'] }
        ]
      },
      {
        type: 'log',
        label: 'Log',
        category: 'Utility',
        parameters: [
          { name: 'message', label: 'Message', type: 'string', required: true }
        ]
      }
    ]
  }).as('getNodeTypes')

  // Mock workflow data endpoints
  cy.intercept('GET', '/api/agents/*/versions/latest', {
    statusCode: 200,
    body: {
      nodes: [],
      edges: []
    }
  }).as('getWorkflowVersion')

  cy.intercept('GET', '/api/agents/*/draft', {
    statusCode: 404,
    body: { error: 'No draft found' }
  }).as('getDraft')
})

// Login command (if authentication is needed)
Cypress.Commands.add('login', (username = 'test@example.com', password = 'password') => {
  // Implement login logic when authentication is added
  cy.log('Login command - to be implemented when auth is added')
})

// Visit workflow builder with common setup
Cypress.Commands.add('visitBuilder', (agentId = 'test-agent-1') => {
  cy.mockAPIEndpoints()
  cy.visit(`/builder/${agentId}`)
  cy.wait(['@getNodeTypes', '@getWorkflowVersion'])
})

// Drag and drop node from sidebar
Cypress.Commands.add('addNodeToCanvas', (nodeType, position = { x: 300, y: 200 }) => {
  // Find the node in the sidebar and drag it to the canvas
  cy.get(`[data-node-type="${nodeType}"]`)
    .trigger('dragstart')
  
  cy.get('[data-testid="workflow-canvas"]')
    .trigger('dragover', { clientX: position.x, clientY: position.y })
    .trigger('drop', { clientX: position.x, clientY: position.y })
})

// Connect two nodes
Cypress.Commands.add('connectNodes', (sourceNodeId, targetNodeId) => {
  // Click on source node output handle
  cy.get(`[data-nodeid="${sourceNodeId}"] .react-flow__handle-right`)
    .trigger('mousedown', { button: 0 })
  
  // Drag to target node input handle
  cy.get(`[data-nodeid="${targetNodeId}"] .react-flow__handle-left`)
    .trigger('mousemove')
    .trigger('mouseup')
})

// Test node configuration
Cypress.Commands.add('configureNode', (nodeId, config) => {
  // Click on node to open configuration
  cy.get(`[data-nodeid="${nodeId}"]`).click()
  
  // Fill in configuration based on the config object
  Object.entries(config).forEach(([key, value]) => {
    cy.get(`[name="${key}"]`).clear().type(value)
  })
  
  // Save configuration
  cy.get('[data-testid="save-node-config"]').click()
})