<?php
    $filename = "/pwned/pwned-passwords-sha1-ordered-by-hash-v7.bin";
    $count = filesize($filename) / 20;
    if ($count == 0) {
        http_response_code(500);  // no hashes found
        return;
    }

    $path = $_SERVER['REQUEST_URI'];
    $lastPath = ltrim(end(explode("/", $path)), "?");
    if (!preg_match("/^[0-9A-Fa-f]{5}$/", $lastPath)) {
        http_response_code(400);  // hex match
        return;
    }

    $prefix = hexdec($lastPath);
    if (($prefix < 0) || ($prefix > 1048575)) {
        http_response_code(400);  // wrong prefix
        return;
    }


    $file = fopen($filename, "rb");

    $output = "";
    $min = findFirstMatch($file, $prefix, $count);
    for ($i = $min; $i < $count; $i++) {
        $buffer = readHashAt($file, $i);
	$currPrefix = getPrefix($buffer);
	if ($currPrefix != $prefix) { break; }
        $output .= bin2hex($buffer) . "\n";
    }

    fclose($file);


    header("Content-Type: text/plain");
    echo $output;


    function findFirstMatch($file, int $prefix, int $count) {
        $index = findMatch($file, $prefix, 0, $count - 1);
        while ($index > 0) {
            $prevPrefix = readHashPrefixAt($file, $index - 1);
            if ($prevPrefix != $prefix) { break; }
            $index -= 1;
        }
        return $index;
    }

    function findMatch($file, int $prefix, int $minIndex, int $maxIndex) {
        if ($minIndex > $maxIndex) { return -1; }
	$pivot = intdiv($minIndex + $maxIndex, 2);
        $currPrefix = readHashPrefixAt($file, $pivot);
        if ($currPrefix == $prefix) {
            return $pivot;
        } else if ($prefix < $currPrefix) {
            return findMatch($file, $prefix, $minIndex, $pivot - 1);
        } else {
            return findMatch($file, $prefix, $pivot + 1, $maxIndex);
        }
    }

    function readHashPrefixAt($file, int $index) {
        $buffer = readHashAt($file, $index);
        $prefix = ord(substr($buffer, 0))<<12 | ord(substr($buffer, 1))<<4 | ord(substr($buffer, 2))>>4;
        return $prefix;
    }

    function readHashAt($file, int $index) {
        $offset = $index * 20;
	fseek($file, $offset, SEEK_SET);
	$buffer = fread($file, 20);
        return $buffer;
    }

    function getPrefix(string $buffer) {
        return ord(substr($buffer, 0))<<12 | ord(substr($buffer, 1))<<4 | ord(substr($buffer, 2))>>4;
    }

?>
