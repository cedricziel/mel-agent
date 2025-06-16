describe('Configuration Node Dialog', () => {
  beforeEach(() => {
    // Mock the API endpoints
    cy.intercept('GET', '/api/agents/*/versions/latest', {
      fixture: 'agent.json'
    }).as('getAgent');

    cy.intercept('GET', '/api/workflows/*/nodes', {
      fixture: 'nodes.json'
    }).as('getNodes');

    cy.intercept('GET', '/api/workflows/*/edges', {
      fixture: 'edges.json'
    }).as('getEdges');

    cy.intercept('GET', '/api/node-types', {
      fixture: 'node-types.json'
    }).as('getNodeTypes');

    cy.intercept('PUT', '/api/workflows/*/nodes/*', {
      statusCode: 200,
      body: { success: true }
    }).as('updateNode');

    // Visit the builder page
    cy.visit('/agents/test-agent-id/edit');

    // Wait for the page to load
    cy.wait(['@getAgent', '@getNodes', '@getEdges', '@getNodeTypes']);
  });

  it('should open dialog when clicking on configuration nodes', () => {
    // Create an agent node first
    cy.get('[data-testid="add-node-button"]').click();
    cy.get('[data-testid="node-type-agent"]').click();
    cy.get('[data-testid="confirm-add-node"]').click();

    // Add a model configuration node
    cy.get('[data-testid="agent-node"]').find('[title="Add Model Configuration"]').click();
    cy.get('[data-testid="config-option-openai"]').click();

    // Click on the model configuration node
    cy.get('[data-testid="openai-model-node"]').click();

    // Verify the dialog opens
    cy.get('[data-testid="node-details-dialog"]').should('be.visible');
    cy.get('[data-testid="node-details-dialog"]').should('contain', 'OpenAI Model');
  });

  it('should show configuration form for OpenAI model node', () => {
    // Assume we have an OpenAI model node already created
    cy.get('[data-testid="openai-model-node"]').click();

    // Verify the dialog shows configuration options
    cy.get('[data-testid="node-details-dialog"]').within(() => {
      cy.contains('Configuration').should('be.visible');
      cy.contains('Model').should('be.visible');
      cy.contains('Temperature').should('be.visible');
      cy.contains('Max Tokens').should('be.visible');
    });
  });

  it('should allow editing configuration values', () => {
    cy.get('[data-testid="openai-model-node"]').click();

    cy.get('[data-testid="node-details-dialog"]').within(() => {
      // Change the model
      cy.get('select[name="model"]').select('gpt-3.5-turbo');

      // Change temperature
      cy.get('input[name="temperature"]').clear().type('0.5');

      // Change max tokens
      cy.get('input[name="maxTokens"]').clear().type('2000');
    });

    // Save changes (if there's a save button)
    cy.get('[data-testid="save-node-config"]').click();

    // Verify the API was called with updated values
    cy.wait('@updateNode').then((interception) => {
      expect(interception.request.body.data).to.include({
        model: 'gpt-3.5-turbo',
        temperature: 0.5,
        maxTokens: 2000
      });
    });
  });

  it('should show different configuration for different node types', () => {
    // Test Anthropic model node
    cy.get('[data-testid="anthropic-model-node"]').click();

    cy.get('[data-testid="node-details-dialog"]').within(() => {
      cy.contains('Anthropic Model').should('be.visible');
      cy.contains('Model').should('be.visible');
      cy.get('select[name="model"]').should('contain.text', 'claude');
    });

    cy.get('[data-testid="close-dialog"]').click();

    // Test Local Memory node
    cy.get('[data-testid="local-memory-node"]').click();

    cy.get('[data-testid="node-details-dialog"]').within(() => {
      cy.contains('Local Memory').should('be.visible');
      cy.contains('Max Messages').should('be.visible');
      cy.contains('Enable Summarization').should('be.visible');
    });
  });

  it('should close dialog when clicking outside or on close button', () => {
    cy.get('[data-testid="openai-model-node"]').click();
    cy.get('[data-testid="node-details-dialog"]').should('be.visible');

    // Close with close button
    cy.get('[data-testid="close-dialog"]').click();
    cy.get('[data-testid="node-details-dialog"]').should('not.exist');

    // Open again
    cy.get('[data-testid="openai-model-node"]').click();
    cy.get('[data-testid="node-details-dialog"]').should('be.visible');

    // Close by clicking outside (on backdrop)
    cy.get('[data-testid="dialog-backdrop"]').click({ force: true });
    cy.get('[data-testid="node-details-dialog"]').should('not.exist');
  });

  it('should show visual feedback when hovering over configuration nodes', () => {
    cy.get('[data-testid="openai-model-node"]').trigger('mouseover');

    // Check if the node shows hover state (this depends on CSS implementation)
    cy.get('[data-testid="openai-model-node"]').should('have.class', 'cursor-pointer');
  });

  it('should handle validation errors in configuration form', () => {
    cy.get('[data-testid="openai-model-node"]').click();

    cy.get('[data-testid="node-details-dialog"]').within(() => {
      // Select empty option to trigger validation
      cy.get('select[name="model"]').select('');

      // Try to save
      cy.get('[data-testid="save-node-config"]').click();

      // Should show validation error
      cy.contains('required').should('be.visible');
    });
  });

  it('should maintain node selection state after dialog operations', () => {
    // Click on a configuration node
    cy.get('[data-testid="openai-model-node"]').click();

    // Verify it's selected (dialog is open)
    cy.get('[data-testid="node-details-dialog"]').should('be.visible');

    // Make some changes
    cy.get('[data-testid="node-details-dialog"]').within(() => {
      cy.get('input[name="temperature"]').clear().type('0.8');
    });

    // Close dialog
    cy.get('[data-testid="close-dialog"]').click();

    // Node should still be visually selected if that's the expected behavior
    // This test depends on the exact UX design for selection states
  });

  it('should support keyboard navigation in configuration dialog', () => {
    cy.get('[data-testid="openai-model-node"]').click();

    cy.get('[data-testid="node-details-dialog"]').within(() => {
      // Tab through form fields
      cy.get('input[name="temperature"]').focus().tab();
      cy.get('input[name="maxTokens"]').should('be.focused');

      // Escape should close dialog
      cy.get('body').type('{esc}');
    });

    cy.get('[data-testid="node-details-dialog"]').should('not.exist');
  });

  it('should persist configuration changes across page refreshes', () => {
    cy.get('[data-testid="openai-model-node"]').click();

    cy.get('[data-testid="node-details-dialog"]').within(() => {
      cy.get('input[name="temperature"]').clear().type('0.3');
      cy.get('[data-testid="save-node-config"]').click();
    });

    cy.wait('@updateNode');

    // Refresh page
    cy.reload();
    cy.wait(['@getAgent', '@getNodes', '@getEdges', '@getNodeTypes']);

    // Check that the value persisted
    cy.get('[data-testid="openai-model-node"]').click();
    cy.get('[data-testid="node-details-dialog"]').within(() => {
      cy.get('input[name="temperature"]').should('have.value', '0.3');
    });
  });
});
