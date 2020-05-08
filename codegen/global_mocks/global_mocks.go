package global_mocks

/************************************************************************************************
*
*
*        This file is used as a stub that only contains a go:generate mockgen statement
*        that triggers a shell script that generates external library mocks in parallel.
*
*                 NOTE: Do not remove the rest of this line: `go:generate mockgen`
*                       This is how we ensure this file is run when we do `make mockgen`
*
*
************************************************************************************************/

//go:generate ./parallel_mockgen.sh
