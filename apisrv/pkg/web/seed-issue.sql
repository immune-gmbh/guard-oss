INSERT INTO v2.organizations (id, external, devices, features, updated_at)
VALUES (
        100,
        'ext-id-1',
        100,
        ARRAY []::v2.organizations_feature [],
        'NOW'
    );


INSERT INTO v2.devices (
        id,
        hwid,
        fpr,
        name,
        retired,
        organization_id,
        replaced_by,
        baseline
    )
VALUES (
        100,
        E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        'Test Device #1',
        FALSE,
        100,
        NULL,
        '{"type": "dummy"}'
    ),
    (
        101,
        E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234b',
        'Test Device #2',
        TRUE,
        100,
        100,
        '{"type": "dummy"}'
    ),
    (
        102,
        E'\\x0022000b305c1823252de4490e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        E'\\x0022000b305c1823253de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        'Test Device #3',
        FALSE,
        100,
        NULL,
        '{"type": "dummy"}'
    ),
    (
        103,
        E'\\x0022000b305c1823253de4490e640ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        E'\\x0022000b305c1823256de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        'Test Device #4',
        FALSE,
        100,
        NULL,
        '{"type": "dummy"}'
    );


INSERT INTO v2.keys (id, public, name, fpr, credential, device_id)
VALUES (
        100,
        E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251',
        'aik',
        E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        '{}',
        100
    ),
    (
        102,
        E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251',
        'aik',
        E'\\x0022000b305c1823255de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        '{}',
        102
    ),
    (
        103,
        E'\\x0001000b000300b20020837197674484b3f81a90cc8d46a5d724fd52d76e06520b64f2a1da1b331469aa00060080004300100800000000000100d2b2c1d0aa3955801bd01785b2d51a9ecdf3acb016e43eeaa74c5515173a4447258c26ac56b03d090bbbb091fe1ab05c2209c5f0e34eb6a0456bada327b524b1c109c671a895d52545017ed3886d13adc9f6d5ffb80b815d7f23eb6e45f3477c28cc4f048cfc0f11e291f706a97b5df7dc7732397fc723f187ea2d74d69afe7cd09e4677de14533737eb9b3942147fca236c1db29f97539bbd208d6ac0ae41ba2d99f1f7a56cd5a478e8ed9aa1a8d716fb05e47bea5a081ad834f9998564e885acfbd5d7a55d83740e17fb6b11b0ecd308dc590b9365797eb147b392f8d491d37fc2d21e0df40391047d23700c5da72abaa0f5863471184f0d7307edb72e4251',
        'aik',
        E'\\x0022000b305c1823255de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec34c1234a',
        '{}',
        103
    );


INSERT INTO v2.evidence (
        id,
        received_at,
        signed_by,
        VALUES,
        baseline,
        device_id
    )
VALUES (
        '2AWyQLT8uKgyTnq2MB1VtO99hGY',
        'NOW',
        E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        '{"type": "dummy"}',
        '{"type": "dummy"}',
        100
    ),
    (
        '2AWyQNZPqzzjX4zNZGDw6zvThuG',
        'NOW',
        E'\\x0022000b305c1823255de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        '{"type": "dummy"}',
        '{"type": "dummy"}',
        102
    ),
    (
        '2AWyQNrDAhfQq2cWlsxIWsdAfFg',
        'NOW',
        E'\\x0022000b305c1823252de4490e639ec0c327f8d1da14e18634ef7f6c3b57b0ec33c1234a',
        '{"type": "dummy"}',
        '{"type": "dummy"}',
        103
    );


INSERT INTO v2.appraisals (
        id,
        device_id,
        received_at,
        appraised_at,
        expires,
        verdict,
        evidence_id,
        report,
        key_id
    )
VALUES (
        100,
        100,
        'NOW',
        'NOW',
        '3022-10-10 00:00:00 UTC',
        '{"type": "verdict/3", "result": "trusted", "firmware": "trusted", "bootloader": "trusted", "supply_chain": "unsupported", "configuration": "trusted", "operating_system": "trusted", "endpoint_protection": "unsupported"}'::jsonb,
        '2AWyQLT8uKgyTnq2MB1VtO99hGY',
        '{"type": "report/1"}',
        100
    ),
    (
        102,
        102,
        'NOW',
        'NOW',
        '3022-10-10 00:00:00 UTC',
        '{"type": "verdict/3", "result": "trusted", "firmware": "trusted", "bootloader": "trusted", "supply_chain": "unsupported", "configuration": "trusted", "operating_system": "trusted", "endpoint_protection": "unsupported"}'::jsonb,
        '2AWyQNZPqzzjX4zNZGDw6zvThuG',
        '{"type": "report/1"}',
        102
    ),
    (
        103,
        103,
        timestamp 'NOW' - INTERVAL '2 days',
        timestamp 'NOW' - INTERVAL '2 days',
        timestamp 'NOW' - INTERVAL '1 day',
        '{"type": "verdict/3", "result": "trusted", "firmware": "trusted", "bootloader": "trusted", "supply_chain": "unsupported", "configuration": "trusted", "operating_system": "trusted", "endpoint_protection": "unsupported"}'::jsonb,
        '2AWyQNrDAhfQq2cWlsxIWsdAfFg',
        '{"type": "report/1"}',
        103
    );


INSERT INTO v2.issues_appraisals (
        appraisal_id,
        issue_type,
        incident,
        payload
    )
VALUES (100, 'issue1', false, '{"type": "dummy"}'),
    (102, 'issue1', false, '{"type": "dummy"}'),
    (103, 'issue2', false, '{"type": "dummy"}');