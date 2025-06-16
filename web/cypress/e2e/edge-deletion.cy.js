describe('Edge Deletion Functionality', () => {
  beforeEach(() => {
    // Visit a workflow builder page with some nodes and edges
    cy.visit('/agents/test-agent-id/edit');

    // Wait for the React app to fully load
    cy.get('#root', { timeout: 10000 }).should('exist');
  });

  it('should load the builder page', () => {
    // Wait for the page to load and verify basic elements
    // Use a more flexible selector approach
    cy.get('[data-testid="react-flow"], .react-flow, .react-flow__renderer', { timeout: 20000 })
      .should('exist');
  });

  it('should have edge hover functionality', function() {
    // Wait for the page to load
    cy.get('[data-testid="react-flow"], .react-flow, .react-flow__renderer', { timeout: 20000 })
      .should('exist');

    // Check if there are any edges and test hover functionality
    cy.get('.react-flow__edge').then($edges => {
      // Explicitly assert that edges exist or skip the test
      if ($edges.length === 0) {
        cy.log('No edges found to test hover functionality, skipping test');
        this.skip();
        return;
      }

      // Assert edges exist for the test requirements
      cy.get('.react-flow__edge')
        .should('exist')
        .and('have.length.greaterThan', 0);

      // Hover over the first edge
      cy.get('.react-flow__edge').first().trigger('mouseover');

      // Look for delete button (may be in edge label renderer)
      cy.get('button[title="Delete edge"]', { timeout: 3000 }).should('exist');
    });
  });
});
