#include "estatistica.h"

#include <algorithm>
#include <vector>

double mediana(const double *v, int n) {
	std::vector<double> c(v, v + n);
	std::sort(c.begin(), c.end());
	if (n % 2 == 1) return c[n / 2];
	return (c[n / 2 - 1] + c[n / 2]) / 2;
}
