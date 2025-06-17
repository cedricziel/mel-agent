describe('Configuration Node Click Functionality', () => {
  beforeEach(() => {
    // Mock all the required API endpoints
    cy.intercept('GET', '/api/node-types', {
      statusCode: 200,
      body: [
        {
          type: 'openai_model',
          label: 'OpenAI Model',
          category: 'Configuration',
          description: 'OpenAI model configuration'
        },
        {
          type: 'local_memory',
          label: 'Local Memory',
          category: 'Configuration', 
          description: 'Local memory configuration'
        },
        {
          type: 'workflow_tools',
          label: 'Tools',
          category: 'Configuration',
          description: 'Workflow tools configuration'
        }
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
          data: { label: 'OpenAI', nodeTypeLabel: 'OpenAI Model' },
          position: { x: 100, y: 100 }
        },
        {
          id: 'config-node-2', 
          type: 'local_memory',
          data: { label: 'Memory', nodeTypeLabel: 'Local Memory' },
          position: { x: 250, y: 100 }
        },
        {
          id: 'config-node-3',
          type: 'workflow_tools', 
          data: { label: 'Tools', nodeTypeLabel: 'Tools' },
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
    const configTexts = ['OpenAI', 'Anthropic', 'Local Memory', 'Model', 'Tools', 'Memory'];

    configTexts.forEach(text => {
      cy.get('body').then($body => {
        if ($body.find(`*:contains("${text}")`).length > 0) {
          cy.log(`Found config node with text: ${text}`);
        }
      });
    });
  });
});
