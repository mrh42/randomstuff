#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <assert.h>
#include <gmp.h>

// results from mpn_get_str() need post-processing.
char *asciify(char *buf, mp_limb_t *n, int len) {
	int i = mpn_get_str(buf, 10,  n, len);
	char *s = buf;
	char *t = buf;
	while (i--) {		
		*s += '0';
		s++;
	}
	while (*t == '0') {
		t++;
	}
	return t;
}
int topbit(uint64_t p) {
	int top = 0;
	uint64_t one = 1l;
	for (int i = 62; i >= 0; i--) {
		if (p & (one << i)) {
			top = i;
			break;
		}
	}
	return top;
}

//
// input: p, k
// output: Q
// returns 1 if a factor was found.
//
int tf(uint64_t p, uint64_t k, mp_limb_t Q[2]) {
		
	int top = topbit(p);
	mp_limb_t One = 1;

	mp_limb_t P[1] = {p};
	mp_limb_t K[1] = {k};
	//mp_limb_t Q[2];

	// Q = P * K * 2 + 1
	mpn_mul_n(Q, P, K, 1);
	mpn_lshift(Q, Q, 2, 1);
	mpn_add_1(Q, Q, 2, One);

	// mpn_tdiv_qr() requires a non-zero most significant limb
	int qn = 2;  // 128 bit?
	if (Q[1] == 0) {
		// only 64-bit
		qn = 1;
	}
	// will be 2=128-bit or 4=256-bit.
	int qn2 = qn << 1;

	//
	// Starting with 1, repeatedly square, remove the top bit of the exponent and if 1 multiply
	// squared value by 2, then compute the remainder
	//
	uint64_t one = 1l;
	mp_limb_t Sq[4] = {1, 0, 0, 0};
	mp_limb_t Sq2[4] = {1, 0, 0, 0};
	for (int b = top; b >= 0; b--) {
		// Sq *= Sq
		mpn_mul_n(Sq2, Sq, Sq, qn2);
		if (p & (one << b)) {
			// Sq *= 2
			mpn_lshift(Sq, Sq2, qn2, 1);
		} else {
			Sq[0] = Sq2[0];
			Sq[1] = Sq2[1];
			Sq[2] = Sq2[2];
			Sq[3] = Sq2[3];
		}
		// Sq %= Q
		mp_limb_t X[4], R[4];
		mpn_tdiv_qr(X, R, 0, Sq, qn2, Q, qn);
		Sq[0] = R[0];
		if (qn == 1) {
			Sq[1] = 0;
		} else {
			Sq[1] = R[1];
			Sq[2] = 0;
			Sq[3] = 0;
		}
	}

	// If Sq == 1, found a factor
	if (Sq[0] == 1) {
		for (int i = 1; i < qn2; i++) {
			if (Sq[i] != 0) {
				return 0;
			}
		}
		return 1;
	}
	return 0;
}

// Test for (3,5)mod8, and (0)mod(primes 3 -> 23) in one shot. From 446,185,740 potential K-values,
// a list of 72,990,720 are left to TF test. 
// When the entire list has been tested, K-base += M.

#define M (4 * 3L * 5 * 7 * 11 * 13 * 17 * 19 * 23)  // 446185740
#define ListLen 72990720
static uint64_t List[ListLen];

void initlist(uint64_t p) {
	int i = 0;
	for (uint64_t k = 0; k < M; k++) {
		uint64_t q = p * k * 2 + 1;
		if (((q&7) == 3) || ((q&7) == 5) || (q%3 == 0) || (q%5 == 0) || (q%7 == 0) ||
		    (q%11 == 0) || (q%13 == 0) || (q%17 == 0) || (q%19 == 0) || (q%23 == 0)) {
			; // nothing
		} else {
			List[i++] = k;
		}
	}
	assert(i == ListLen);
	//printf("l: %d ll: %d\n", i, ListLen);
}

int main(int ac, char **av) {

	uint64_t p = 726064763;
	uint64_t kbase = 1;

	if (ac > 2) {
		p = atol(av[1]);
		kbase = atol(av[2]);
	}
	mp_limb_t Q[2];
	//if (tf(p, kbase, Q)) {
	//	printf("%ld kfactor %ld\n", p, kbase);
	//}
	//return 0;

	// Start with kbase == 0 mod M.
	kbase = (kbase/M) * M;
	initlist(p);
	while (1) {
		printf("kbase: %ld\n", kbase);

		for (int i = 0; i < ListLen; i++) {
			uint64_t o = kbase + List[i];
			if (o && tf(p, o, Q)) {
				char buf[64];
				char *q = asciify(buf, Q, 2);
				printf("%ld kfactor %ld == %s\n", p, o, q);
			}
		}
		// advance to the next chunk of M factors.
		kbase += M;
		//break;  // only run one check.
	}
}
