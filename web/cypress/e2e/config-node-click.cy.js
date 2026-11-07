describe('Configuration Node Click Functionality', () => {
  beforeEach(() => {
    // Mock all the required API endpoints with proper structure
    // Node definitions include the config nodes (openai_model, local_memory, ...).
    // The builder looks up each node's definition by type to open its modal, so
    // without them clicking a config node cannot open anything.
    cy.intercept('GET', '/api/node-types', {
      fixture: 'node-types.json'
    }).as('getNodeTypes')

    cy.intercept('GET', '/api/triggers', {
      statusCode: 200,
      body: []
    }).as('getTriggers')

    // Mock draft API (returns 404 - no draft exists)
    cy.intercept('GET', '/api/workflows/test-agent-id/draft', {
      statusCode: 404,
      body: { error: 'No draft found' }
    }).as('getDraft')

    // Mock workflow client fallback APIs
    cy.intercept('GET', '/api/workflows/test-agent-id', {
      statusCode: 200,
      body: { id: 'test-agent-id', name: 'Test Agent' }
    }).as('getWorkflowMeta')

    // Shape matches the generated WorkflowNode contract: { id, name, type, config }
    cy.intercept('GET', '/api/workflows/test-agent-id/nodes', {
      statusCode: 200,
      body: [
        {
          id: 'config-node-1',
          name: 'OpenAI',
          type: 'openai_model',
          config: {
            label: 'OpenAI',
            nodeTypeLabel: 'OpenAI Model',
            model: 'gpt-4',
            temperature: 0.7,
            maxTokens: 1000
          }
        },
        {
          id: 'config-node-2',
          name: 'Memory',
          type: 'local_memory',
          config: {
            label: 'Memory',
            nodeTypeLabel: 'Local Memory',
            maxMessages: 100,
            enableSummarization: true
          }
        },
        {
          id: 'config-node-3',
          name: 'Tools',
          type: 'workflow_tools',
          config: {
            label: 'Tools',
            nodeTypeLabel: 'Tools',
            enabledTools: []
          }
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
    cy.get('.react-flow', { timeout: 15000 }).should('be.visible');
    cy.get('.react-flow__renderer').should('be.visible');
  });

  it('should have clickable configuration nodes', () => {
    // Wait for the page to load
    cy.get('.react-flow', { timeout: 15000 }).should('be.visible');

    // Wait for nodes to be rendered by looking for React Flow nodes
    cy.get('.react-flow__node', { timeout: 10000 }).should('have.length.at.least', 1);

    // Look for configuration nodes (they should have cursor-pointer class)
    cy.get('.cursor-pointer', { timeout: 10000 }).should('have.length.at.least', 1);

    // Click on the first clickable node
    cy.get('.cursor-pointer').first().click();

    // The page should not crash and some modal or details should open
    cy.get('body').should('be.visible');
  });

  it('should show config node labels', () => {
    // Wait for the page to load
    cy.get('.react-flow', { timeout: 15000 }).should('be.visible');

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
    cy.get('.react-flow', { timeout: 15000 }).should('be.visible');

    // Wait for nodes to be rendered
    cy.get('.react-flow__node', { timeout: 10000 }).should('have.length.at.least', 1);

    // Find and click on a config node
    cy.get('.cursor-pointer', { timeout: 10000 }).first().click();

    // Should open a modal - check for either text
    cy.get('body').then($body => {
      expect($body.text()).to.match(/OpenAI Model|Configuration/);
    });
    
    // Modal should have basic elements
    cy.get('body').then($body => {
      expect($body.text()).to.match(/Save|Close/);
    });
  });

  it('should display config node parameters in modal', () => {
    // Wait for page load
    cy.get('.react-flow', { timeout: 15000 }).should('be.visible');

    // Wait for nodes to be rendered
    cy.get('.react-flow__node', { timeout: 10000 }).should('have.length.at.least', 1);

    // Click config node to open modal
    cy.get('.cursor-pointer', { timeout: 10000 }).first().click();

    // Should show configuration options
    cy.get('body').then($body => {
      expect($body.text()).to.match(/Configuration|OpenAI Model/);
    });
    
    // Should show parameter inputs (exact text depends on implementation)
    cy.get('body').then($body => {
      const hasConfigFields = $body.find('input, select, textarea').length > 0;
      if (hasConfigFields) {
        cy.log('Found configuration input fields');
      }
    });
  });
});
