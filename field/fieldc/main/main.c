#include <stdio.h>
#include "../generic.h"

void main() {
	Element a;
	a.l0 = 2;
	a.l1 = 2;
	a.l2 = 2;
	a.l3 = 2;
	a.l4 = 2;

	feMulGeneric(&a, &a, &a);
	feSquareGeneric(&a, &a);

	printf("%llu %llu %llu %llu %llu\n", a.l0, a.l1, a.l2, a.l3, a.l4);
}