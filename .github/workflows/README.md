== UNIT TESTS ==

Unit tests are optimized to cache their results run only when the paths that relate to their
tested units are touched. However, to be able to use branch protections that require the
unit tests to be run there must be same-name surrogate dummy jobs that are run instead of
the skipped unit tests. These dummy tests are configured to only run when paths are touched
that are not covered by their non-dummy siblings. This solution has the downside that when
all conditions are satisfied, i.e. multiple paths are touched and some are inside and outside
of the paths covered by the unit tests of a workflow, than the dummy and non-dummy workflows
are both run. Luckily the branch protection feature will wait for all of them to complete.