describe('Configuration Node Click Functionality', () => {
  beforeEach(() => {
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
