package rqlimit

type Policy interface {
	Name() string
	Budgets() []Budget
	Costs() []Cost
}

type SimplePolicy struct {
	name    string
	costs   []Cost
	budgets []Budget
}

var _ Policy = &SimplePolicy{}

func (c *SimplePolicy) WithName(n string) *SimplePolicy {
	c.name = n
	return c
}

func (c *SimplePolicy) WithBudgets(budgets ...Budget) *SimplePolicy {
	c.budgets = budgets
	return c
}

func (c *SimplePolicy) WithCosts(costs ...Cost) *SimplePolicy {
	c.costs = costs
	return c
}

func (c *SimplePolicy) Name() string      { return c.name }
func (c *SimplePolicy) Budgets() []Budget { return c.budgets }
func (c *SimplePolicy) Costs() []Cost     { return c.costs }
