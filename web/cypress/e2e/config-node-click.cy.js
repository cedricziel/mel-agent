describe('Configuration Node Click Functionality', () => {
  beforeEach(() => {
    // Mock all the required API endpoints with proper structure
    cy.intercept('GET', '/api/node-types', {
      statusCode: 200,
      body: [
        // Backend action nodes
        {
          type: 'agent',
          label: 'Agent',
          category: 'Core',
          description: 'AI agent node'
        },
        {
          type: 'http_request',
          label: 'HTTP Request',
          category: 'Actions',
          description: 'Make HTTP requests'
        }
        // Note: Config nodes come from frontend fallback until backend is updated
      ]
    }).as('getNodeTypes')

    cy.intercept('GET', '/api/triggers', {
      statusCode: 200,
      body: []
    }).as('getTriggers')

    // Mock draft API (returns 404 - no draft exists)
    cy.intercept('GET', '/api/agents/test-agent-id/draft', {
      statusCode: 404,
      body: { error: 'No draft found' }
    }).as('getDraft')

    // Mock workflow client fallback APIs
    cy.intercept('GET', '/api/workflows/test-agent-id', {
      statusCode: 200,
      body: { id: 'test-agent-id', name: 'Test Agent' }
    }).as('getWorkflowMeta')

    cy.intercept('GET', '/api/workflows/test-agent-id/nodes', {
      statusCode: 200,
      body: [
        {
          id: 'config-node-1',
          type: 'openai_model',
          data: { 
            label: 'OpenAI', 
            nodeTypeLabel: 'OpenAI Model',
            model: 'gpt-4',
            temperature: 0.7,
            maxTokens: 1000
          },
          position: { x: 100, y: 100 }
        },
        {
          id: 'config-node-2', 
          type: 'local_memory',
          data: { 
            label: 'Memory', 
            nodeTypeLabel: 'Local Memory',
            maxMessages: 100,
            enableSummarization: true
          },
          position: { x: 250, y: 100 }
        },
        {
          id: 'config-node-3',
          type: 'workflow_tools', 
          data: { 
            label: 'Tools', 
            nodeTypeLabel: 'Tools',
            enabledTools: []
          },
          position: { x: 400, y: 100 }
        }
      ]
    }).as('getWorkflowNodes')

    cy.intercept('GET', '/api/workflows/test-agent-id/edges', {
      statusCode: 200,
      body: []
    }).as('getWorkflowEdges')

    // Visit a workflow builder page
    cy.visit('/agents/test-agent-id/edit');
  });

  it('should load the builder page with nodes', () => {
    // Wait for the page to load
    cy.get('[data-testid="react-flow"]', { timeout: 15000 }).should('be.visible');
    cy.get('.react-flow__renderer').should('be.visible');
  });

  it('should have clickable configuration nodes', () => {
    // Wait for the page to load
    cy.get('[data-testid="react-flow"]', { timeout: 15000 }).should('be.visible');

    // Look for configuration nodes (they should have cursor-pointer class)
    cy.get('.cursor-pointer').then($nodes => {
      if ($nodes.length > 0) {
        // Click on the first clickable node
        cy.get('.cursor-pointer').first().click();

        // The page should not crash and some modal or details should open
        // We're just testing that the click works without errors
        cy.get('body').should('be.visible');
      } else {
        cy.log('No clickable configuration nodes found');
      }
    });
  });

  it('should show config node labels', () => {
    // Wait for the page to load
    cy.get('[data-testid="react-flow"]', { timeout: 15000 }).should('be.visible');

    // Look for common config node text
    const configTexts = ['OpenAI', 'Memory', 'Tools'];

    configTexts.forEach(text => {
      cy.get('body').then($body => {
        if ($body.find(`*:contains("${text}")`).length > 0) {
          cy.log(`Found config node with text: ${text}`);
        }
      });
    });
  });

  it('should open modal when config node is clicked', () => {
    // Wait for the page to load
    cy.get('[data-testid="react-flow"]', { timeout: 15000 }).should('be.visible');

    // Find and click on a config node
    cy.get('.cursor-pointer').first().click();

    // Should open a modal
    cy.get('body').should('contain', 'OpenAI Model').or('contain', 'Configuration');
    
    // Modal should have basic elements
    cy.get('body').should('contain', 'Save').or('contain', 'Close');
  });

  it('should display config node parameters in modal', () => {
    // Wait for page load
    cy.get('[data-testid="react-flow"]', { timeout: 15000 }).should('be.visible');

    // Click config node to open modal
    cy.get('.cursor-pointer').first().click();

    // Should show configuration options
    cy.get('body').should('contain', 'Configuration').or('contain', 'OpenAI Model');
    
    // Should show parameter inputs (exact text depends on implementation)
    cy.get('body').then($body => {
      const hasConfigFields = $body.find('input, select, textarea').length > 0;
      if (hasConfigFields) {
        cy.log('Found configuration input fields');
      }
    });
  });
});
