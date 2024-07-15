/* 
   This code was largely inspired by 
   pyrit opencl kernel sha1 routines.
   Copyright 2011 by chucheng 
   zunceng at gmail dot com
   This program comes with ABSOLUTELY NO WARRANTY; express or
   implied .
   This is free software, and you are welcome to redistribute it
   under certain conditions; as expressed here 
   http://www.gnu.org/licenses/gpl-2.0.html
*/

#ifdef cl_khr_byte_addressable_store
#pragma OPENCL EXTENSION cl_khr_byte_addressable_store : disable
#endif

#ifdef cl_nv_pragma_unroll
#define NVIDIA
#pragma OPENCL EXTENSION cl_nv_pragma_unroll : enable
#endif

#ifdef NVIDIA
inline uint SWAP32(uint x)
{
	x = rotate(x, 16U);
	return ((x & 0x00FF00FF) << 8) + ((x >> 8) & 0x00FF00FF);
}
#else
#define SWAP32(a)	(as_uint(as_uchar4(a).wzyx))
#endif

#define K0  0x5A827999
#define K1  0x6ED9EBA1
#define K2  0x8F1BBCDC
#define K3  0xCA62C1D6

#define H1 0x67452301
#define H2 0xEFCDAB89
#define H3 0x98BADCFE
#define H4 0x10325476
#define H5 0xC3D2E1F0

#ifndef uint32_t
#define uint32_t unsigned int
#endif

uint32_t SHA1CircularShift(int bits, uint32_t word)
{
	return ((word << bits) & 0xFFFFFFFF) | (word) >> (32 - (bits));
}

__kernel void sha1_kernel(__global char *message,  __global uint *digest){
    int t, gid, msg_pad;
    int stop, mmod;
    uint i, item, ulen, total;
    uint W[80], temp, A,B,C,D,E;
	int current_pad;

	int idx = get_global_id(0);
	message += 68*idx;
	digest += 5*idx;

    ulen = *(__global uint *)message;
    message += 4;

	msg_pad = 0;
	total = ulen%64 >= 56 ? 2 : 1 + ulen/64;

    digest[0] = 0x67452301;
	digest[1] = 0xEFCDAB89;
	digest[2] = 0x98BADCFE;
	digest[3] = 0x10325476;
	digest[4] = 0xC3D2E1F0;

	for(item = 0; item < total; item++) {
		A = digest[0];
		B = digest[1];
		C = digest[2];
		D = digest[3];
		E = digest[4];

	#pragma unroll
		for (t = 0; t < 80; t++) {
		    W[t] = 0x00000000;
		}
		msg_pad = item * 64;
		if(ulen > msg_pad) {
			current_pad = (ulen - msg_pad) > 64 ? 64 : (ulen - msg_pad);
		} else {
			current_pad = -1;
		}

		if(current_pad > 0) {
			i=current_pad;

			stop =  i/4;
			for (t = 0 ; t < stop ; t++) {
				W[t] = ((uchar)  message[msg_pad + t * 4]) << 24;
				W[t] |= ((uchar) message[msg_pad + t * 4 + 1]) << 16;
				W[t] |= ((uchar) message[msg_pad + t * 4 + 2]) << 8;
				W[t] |= (uchar)  message[msg_pad + t * 4 + 3];
			}
			mmod = i % 4;
			if (mmod == 3) {
				W[t] = ((uchar)  message[msg_pad + t * 4]) << 24;
				W[t] |= ((uchar) message[msg_pad + t * 4 + 1]) << 16;
				W[t] |= ((uchar) message[msg_pad + t * 4 + 2]) << 8;
				W[t] |=  ((uchar) 0x80);
			} else if (mmod == 2) {
				W[t] = ((uchar)  message[msg_pad + t * 4]) << 24;
				W[t] |= ((uchar) message[msg_pad + t * 4 + 1]) << 16;
				W[t] |=  0x8000;
			} else if (mmod == 1) {
				W[t] = ((uchar)  message[msg_pad + t * 4]) << 24;
				W[t] |=  0x800000;
			} else /*if (mmod == 0)*/ {
				W[t] =  0x80000000;
			}
			
			if (current_pad < 56) {
				W[15] = ulen * 8;
			}
		} else if(current_pad < 0) {
			if (ulen % 64 == 0)
				W[0]=0x80000000;
			W[15] = ulen * 8;
		}

		for (t = 16; t < 80; t++) {
			W[t] = SHA1CircularShift(1, W[t - 3] ^ W[t - 8] ^ W[t - 14] ^ W[t - 16]);
		}

		for (t = 0; t < 20; t++) {
			temp = SHA1CircularShift(5, A) +
				((B & C) | ((~B) & D)) + E + W[t] + K0;
			temp &= 0xFFFFFFFF;
			E = D;
			D = C;
			C = SHA1CircularShift(30, B);
			B = A;
			A = temp;
		}

		for (t = 20; t < 40; t++) {
			temp = SHA1CircularShift(5, A) + (B ^ C ^ D) + E + W[t] + K1;
			temp &= 0xFFFFFFFF;
			E = D;
			D = C;
			C = SHA1CircularShift(30, B);
			B = A;
			A = temp;
		}

		for (t = 40; t < 60; t++) {
			temp = SHA1CircularShift(5, A) +
				((B & C) | (B & D) | (C & D)) + E + W[t] + K2;
			temp &= 0xFFFFFFFF;
			E = D;
			D = C;
			C = SHA1CircularShift(30, B);
			B = A;
			A = temp;
		}

		for (t = 60; t < 80; t++) {
			temp = SHA1CircularShift(5, A) + (B ^ C ^ D) + E + W[t] + K3;
			temp &= 0xFFFFFFFF;
			E = D;
			D = C;
			C = SHA1CircularShift(30, B);
			B = A;
			A = temp;
		}

		digest[0] = (digest[0] + A) & 0xFFFFFFFF;
		digest[1] = (digest[1] + B) & 0xFFFFFFFF;
		digest[2] = (digest[2] + C) & 0xFFFFFFFF;
		digest[3] = (digest[3] + D) & 0xFFFFFFFF;
		digest[4] = (digest[4] + E) & 0xFFFFFFFF;
	}
}