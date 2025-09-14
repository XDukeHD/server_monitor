<?php
header('Content-Type: application/json');
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type');

if ($_SERVER['REQUEST_METHOD'] == 'OPTIONS') {
    exit(0);
}

$host = 'your-database-host'; // Replace with your actual database host
$dbname = 'your-database-name'; // Replace with your actual database name
$username = 'your-database-username'; // Replace with your actual database username
$password = 'your-database-password'; // Replace with your actual database password

try {
    $pdo = new PDO("mysql:host=$host;dbname=$dbname;charset=utf8", $username, $password);
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    $stmt = $pdo->prepare("SELECT id, name, status FROM dedicated_servers WHERE status = 'active'");
    $stmt->execute();
    $servers = $stmt->fetchAll(PDO::FETCH_ASSOC);

    $totalServers = count($servers);
    $serverStatus = [];

    foreach ($servers as $server) {
        $serverId = $server['id'];
        $name = $server['name'];

        $stmt2 = $pdo->prepare("SELECT cpu_usage, ram_usage, disk_usage, server_running_state, collected_at FROM server_stats WHERE server_id = ? ORDER BY collected_at DESC LIMIT 1");
        $stmt2->execute([$serverId]);
        $lastStat = $stmt2->fetch(PDO::FETCH_ASSOC);

        if ($lastStat) {
            $collectedAt = new DateTime($lastStat['collected_at'], new DateTimeZone('UTC'));
            $now = new DateTime('now', new DateTimeZone('UTC'));
            $diff = $now->diff($collectedAt);
            $minutesAgo = ($diff->days * 1440) + ($diff->h * 60) + $diff->i;

            $status = ($minutesAgo > 40) ? 'Unavailable' : 'Running - ' . $lastStat['server_running_state'];
            $cpu = $lastStat['cpu_usage'];
            $ram = $lastStat['ram_usage'];

            $stmt3 = $pdo->prepare("SELECT MIN(collected_at) as first, MAX(collected_at) as last FROM server_stats WHERE server_id = ?");
            $stmt3->execute([$serverId]);
            $uptimeData = $stmt3->fetch(PDO::FETCH_ASSOC);
            $uptimeDays = 0;
            $uptimePercentage = 0;
            if ($uptimeData['first']) {
                $first = new DateTime($uptimeData['first'], new DateTimeZone('UTC'));
                $uptimeDiff = $now->diff($first);
                $uptimeDays = $uptimeDiff->days;
                $expected = $uptimeDays * 480;
                $stmt4 = $pdo->prepare("SELECT COUNT(*) as count FROM server_stats WHERE server_id = ?");
                $stmt4->execute([$serverId]);
                $count = $stmt4->fetch(PDO::FETCH_ASSOC)['count'];
                $uptimePercentage = $expected > 0 ? min(100, ($count / $expected) * 100) : 100;
            }

            $isRunning = 'stopped';
            $loadStatus = '';
            if ($lastStat) {
                if ($minutesAgo <= 40) {
                    $isRunning = 'running';
                    $loadStatus = $lastStat['server_running_state'];
                } else {
                    $isRunning = 'unavailable';
                    $loadStatus = $lastStat['server_running_state'];
                }
            }

            $serverStatus[$serverId] = [
                'name' => $name,
                'server_isRunning' => $isRunning,
                'server_status' => $loadStatus,
                'cpu' => $cpu,
                'ram' => $ram,
                'uptime_days' => $uptimeDays,
                'uptime_percentage' => round($uptimePercentage, 2)
            ];
        } else {
            $serverStatus[$serverId] = [
                'name' => $name,
                'server_isRunning' => 'stopped',
                'server_status' => '',
                'cpu' => 0,
                'ram' => 0,
                'uptime_days' => 0,
                'uptime_percentage' => 0
            ];
        }
    }

    echo json_encode([
        'status' => 'success',
        'totalServers' => $totalServers,
        'server_status' => $serverStatus
    ]);

} catch (PDOException $e) {
    echo json_encode(['status' => 'error', 'message' => 'Database error']);
}
?>
