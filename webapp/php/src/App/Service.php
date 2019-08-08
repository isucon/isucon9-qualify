<?php


namespace App;

use GuzzleHttp\Client;
use PDO;
use Psr\Container\ContainerInterface;
use Psr\Log\LoggerInterface;
use Slim\Http\Request;
use Slim\Http\Response;

class Service
{
    /**
     * @var LoggerInterface
     */
    private $logger;

    /**
     * @var \PDO
     */
    private $dbh;

    /**
     * @var \SlimSession\Helper
     */
    private $session;

    /**
     * @var array
     */
    private $settings;

    private const DATETIME_SQL_FORMAT = 'Y-m-d h:i:s';

    private const ITEM_STATUS_ON_SALE = 'on_sale';
    private const ITEM_STATUS_TRADING = 'trading';
    private const ITEM_STATUS_SOLD_OUT = 'sold_out';
    private const ITEM_STATUS_STOP = 'stop';
    private const ITEM_STATUS_CANCEL = 'cancel';

    private const TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING = 'wait_shipping';
    private const TRANSACTION_EVIDENCE_STATUS_WAIT_DONE = 'wait_done';
    private const TRANSACTION_EVIDENCE_STATUS_DONE = 'done';

    private const SHIPPING_STATUS_INITIAL = 'initial';
    private const SHIPPING_STATUS_WAIT_PICKUP = 'wait_pickup';
    private const SHIPPING_STATUS_SHIPPING = 'shipping';
    private const SHIPPING_STATUS_DONE = 'done';

    private const ISUCARI_API_TOKEN = 'Bearer 75ugk2m37a750fwir5xr-22l6h4wmue1bwrubzwd0';

    private const PAYMENT_SERVICE_ISUCARI_API_KEY = 'a15400e46c83635eb181-946abb51ff26a868317c';
    private const PAYMENT_SERVICE_ISUCARI_SHOP_ID = '11';

    private const MIN_ITEM_PRICE = 100;
    private const MAX_ITEM_PRICE = 1000000;

    private const BUMP_CHARGE_SECONDS = 3;

    private const ITEM_PER_PAGE = 48;

    // constructor receives container instance
    public function __construct(ContainerInterface $container)
    {
        $this->logger = $container->get('logger');
        $this->dbh = $container->get('dbh');
        $this->session = $container->get('session');
        $this->settings = $container->get('settings');
    }

    private function jsonPayload(Request $request)
    {
        $data = json_decode($request->getBody());
        if (JSON_ERROR_NONE !== json_last_error()) {
            throw new \InvalidArgumentException(json_last_error_msg());
        }
        return $data;
    }

    private function getCurrentUser()
    {
        if (! $this->session->exists('user_id')) {
            $this->logger->warning('no session');
            throw new \DomainException('no session');
        }

        $user_id = $this->session->get('user_id');
        $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ?');
        $sth->execute([$user_id]);
        $user = $sth->fetch(PDO::FETCH_ASSOC);

        if ($user === false) {
            $this->logger->warning('not found', ['id' => $user['id']]);
            throw new \DomainException('user not found');
        }

        return $user;
    }

    private function getUserSimpleByID($id)
    {
        $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ?');
        $sth->execute([$id]);
        $user = $sth->fetch(PDO::FETCH_ASSOC);
        if ($user === false) {
            return false;
        }
        return [
          'id' => $user['id'],
            'account_name' => $user['account_name'],
            'num_sell_items' => $user['num_sell_items'],
        ];
    }

    private function getCategoryByID($id)
    {
        $sth = $this->dbh->prepare('SELECT * FROM `categories` WHERE `id` = ?');
        $sth->execute([$id]);
        $category = $sth->fetch(PDO::FETCH_ASSOC);
        if ($category === false) {
            return false;
        }
        if ((int) $category['parent_id'] !== 0) {
            $parent = $this->getCategoryByID($category['parent_id']);
            if ($parent === false) {
                return false;
            }
            $category['parent_category_name'] = $parent['category_name'];
        }
        return $category;
    }

    public function new_items(Request $request, Response $response, array $args)
    {
        $itemId = $request->getParam('item_id', "");
        $createdAt = (int) $request->getParam('created_at', 0);

        try {
            if ($itemId !== "" && $createdAt > 0) {
                // paging
                $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `status` IN (?,?) AND `created_at` <= ? AND `id` < ? ORDER BY `created_at` DESC, `id` DESC LIMIT ?');
                $sth->execute([
                    self::ITEM_STATUS_ON_SALE,
                    self::ITEM_STATUS_SOLD_OUT,
                    (new \DateTime())->setTimestamp($createdAt)->format(self::DATETIME_SQL_FORMAT),
                    $itemId,
                    self::ITEM_PER_PAGE + 1,
                ]);
                $items = $sth->fetchAll(PDO::FETCH_ASSOC);
            } else {
                // 1st page
                $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `status` IN (?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT ?');
                $sth->execute([
                    self::ITEM_STATUS_ON_SALE,
                    self::ITEM_STATUS_SOLD_OUT,
                    self::ITEM_PER_PAGE + 1,
                ]);
                $items = $sth->fetchAll(PDO::FETCH_ASSOC);
            }

            $itemSimples = [];
            foreach ($items as $item) {
                $seller = $this->getUserSimpleByID($item['seller_id']);
                if ($seller === false) {
                    return $response->withStatus(404)->withJson(['error' => 'seller not found']);
                }

                $category = $this->getCategoryByID($item['category_id']);
                if ($category === false) {
                    return $response->withStatus(404)->withJson(['error' => 'category not found']);
                }
                $itemSimples[] = [
                  'id' => $item['id'],
                  'seller_id' => $item['seller_id'],
                  'seller' => $seller,
                  'status' => $item['status'],
                  'name' => $item['name'],
                  'price' => $item['price'],
                  'category_id' => $item['category_id'],
                  'category' => $category,
                  'created_at' => (new \DateTime($item['created_at']))->getTimestamp(),
                ];
            }

            $hasNext = false;
            if (count($itemSimples) > self::ITEM_PER_PAGE) {
                $hasNext = true;
                $itemSimples = array_slice($itemSimples, 0, self::ITEM_PER_PAGE-1);
            }
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }
        return $response->withStatus(200)->withJson(
            [
                'items' => $itemSimples,
                'has_next' => $hasNext
            ]
        );
    }

    public function new_category_items(Request $request, Response $response, array $args)
    {
        $rootCategoryId = $args['id'] ?? 0;
        if ((int) $rootCategoryId === 0) {
            return $response->withStatus(400)->withJson(['error' => 'incorrect category id']);
        }

        $rootCategory = $this->getCategoryByID($rootCategoryId);
        if ($rootCategory === false || (int) $rootCategory['parent_id'] !== 0) {
            return $response->withStatus(404)->withJson('category not found');
        }

        try {
            $sth = $this->dbh->prepare('SELECT id FROM `categories` WHERE parent_id=?');
            $sth->execute([$rootCategoryId]);
            $result = $sth->fetchAll(PDO::FETCH_ASSOC);
            $categoryIds = [];
            foreach ($result as $r) {
                $categoryIds[] = $r['id'];
            }

            $itemId = $request->getParam('item_id');
            $createdAt = (int) $request->getParam('created_at', 0);

            if (!empty($itemId) && $createdAt > 0) {
                // paging
                $in = str_repeat('?,', count($categoryIds) - 1) . '?';
                $sth = $this->dbh->prepare("SELECT * FROM `items` WHERE `status` IN (?,?) AND category_id IN (${in}) AND `created_at` <= ? AND `id` < ? ORDER BY `created_at` DESC, `id` DESC LIMIT ?");
                $sth->execute(array_merge(
                    [self::ITEM_STATUS_ON_SALE, self::ITEM_STATUS_SOLD_OUT],
                    $categoryIds,
                    [
                        (new \DateTime())->setTimestamp($createdAt)->format(self::DATETIME_SQL_FORMAT),
                        $itemId,
                        self::ITEM_PER_PAGE + 1,
                    ]
                ));
                $items = $sth->fetchAll(PDO::FETCH_ASSOC);
            } else {
                // 1st page
                $in = str_repeat('?,', count($categoryIds) - 1) . '?';
                $sth = $this->dbh->prepare("SELECT * FROM `items` WHERE `status` IN (?,?) AND category_id IN (${in}) ORDER BY created_at DESC, id DESC LIMIT ?");
                $sth->execute(array_merge(
                    [self::ITEM_STATUS_ON_SALE, self::ITEM_STATUS_SOLD_OUT],
                    $categoryIds,
                    [self::ITEM_PER_PAGE + 1]
                ));
                $items = $sth->fetchAll(PDO::FETCH_ASSOC);
            }

            $itemSimples = [];
            foreach ($items as $item) {
                $seller = $this->getUserSimpleByID($item['seller_id']);
                if ($seller === false) {
                    return $response->withStatus(404)->withJson(['error' => 'seller not found']);
                }

                $category = $this->getCategoryByID($item['category_id']);
                if ($category === false) {
                    return $response->withStatus(404)->withJson(['error' => 'category not found']);
                }
                $itemSimples[] = [
                    'id' => $item['id'],
                    'seller_id' => $item['seller_id'],
                    'seller' => $seller,
                    'status' => $item['status'],
                    'name' => $item['name'],
                    'price' => $item['price'],
                    'category_id' => $item['category_id'],
                    'category' => $category,
                    'created_at' => (new \DateTime($item['created_at']))->getTimestamp(),
                ];
            }

            $hasNext = false;
            if (count($itemSimples) > self::ITEM_PER_PAGE) {
                $hasNext = true;
                $itemSimples = array_slice($itemSimples, 0, self::ITEM_PER_PAGE-1);
            }
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(200)->withJson(
            [
                'root_category_id' => $rootCategory['id'],
                'root_category_name' => $rootCategory['category_name'],
                'items' => $itemSimples,
                'has_next' => $hasNext
            ]
        );
    }

    public function user_items(Request $request, Response $response, array $args)
    {
        $userId = $args['id'] ?? 0;

        $user = $this->getUserSimpleByID($userId);
        if ($user === false) {
            return $response->withStatus(404)->withJson(['error' => 'user not found']);
        }

        $itemId = $request->getParam('item_id');
        $createdAt = (int) $request->getParam('created_at', 0);
        try {
            if ($itemId !== "" && $createdAt > 0) {
                // paging
                $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `seller_id` = ? AND `status` IN (?,?,?) AND `created_at` <= ? AND `id` < ? ORDER BY `created_at` DESC, `id` DESC LIMIT ?');
                $sth->execute([
                    $user['id'],
                    self::ITEM_STATUS_ON_SALE,
                    self::ITEM_STATUS_TRADING,
                    self::ITEM_STATUS_SOLD_OUT,
                    (new \DateTime())->setTimestamp($createdAt)->format(self::DATETIME_SQL_FORMAT),
                    $itemId,
                    self::ITEM_PER_PAGE + 1,
                ]);
                $items = $sth->fetchAll(PDO::FETCH_ASSOC);
            } else {
                // 1st page
                $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `seller_id` = ? AND `status` IN (?,?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT ?');
                $sth->execute([
                    $user['id'],
                    self::ITEM_STATUS_ON_SALE,
                    self::ITEM_STATUS_TRADING,
                    self::ITEM_STATUS_SOLD_OUT,
                    self::ITEM_PER_PAGE + 1,
                ]);
                $items = $sth->fetchAll(PDO::FETCH_ASSOC);
            }

            $itemSimples = [];
            foreach ($items as $item) {
                $seller = $this->getUserSimpleByID($item['seller_id']);
                if ($seller === false) {
                    return $response->withStatus(404)->withJson(['error' => 'seller not found']);
                }

                $category = $this->getCategoryByID($item['category_id']);
                if ($category === false) {
                    return $response->withStatus(404)->withJson(['error' => 'category not found']);
                }
                $itemSimples[] = [
                    'id' => $item['id'],
                    'seller_id' => $item['seller_id'],
                    'seller' => $seller,
                    'status' => $item['status'],
                    'name' => $item['name'],
                    'price' => $item['price'],
                    'category_id' => $item['category_id'],
                    'category' => $category,
                    'created_at' => (new \DateTime($item['created_at']))->getTimestamp(),
                ];
            }

            $hasNext = false;
            if (count($itemSimples) > self::ITEM_PER_PAGE) {
                $hasNext = true;
                $itemSimples = array_slice($itemSimples, 0, self::ITEM_PER_PAGE-1);
            }
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }
        return $response->withStatus(200)->withJson(
            [
                'user' => $user,
                'items' => $itemSimples,
                'has_next' => $hasNext
            ]
        );
    }

    public function register(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(400)->withJson(['error' => 'json decode error']);
        }

        if (empty($payload->account_name) || empty($payload->address) || empty($payload->password)) {
            return $response->withStatus(500)->withJson(['error' => 'all parameters are required']);
        }

        $hashedPassword = password_hash($payload->password, PASSWORD_BCRYPT, ['cost' => 10]);
        if ($hashedPassword === false) {
            return $response->withStatus(500)->withJson(['error' => 'error']);
        }

        try {
            $sth = $this->dbh->prepare('INSERT INTO `users` (`account_name`, `hashed_password`, `address`, `num_sell_items`) VALUES (?, ?, ?, ?)');
            $sth->execute([$payload->account_name, $hashedPassword, $payload->address, 0]);
            $userId = $this->dbh->lastInsertId();
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        return $response->withJson(['id' => $userId, 'account_name' => $payload->account_name, 'address' => $payload->address]);
    }

    public function login(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(400)->withJson(['error' => 'json decode error']);
        }

        if (empty($payload->account_name) || empty($payload->password)) {
            return $response->withStatus(500)->withJson(['error' => 'all parameters are required']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `account_name` = ?');
            $sth->execute([$payload->account_name]);
            $user = $sth->fetch(PDO::FETCH_ASSOC);
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        if (! password_verify($payload->password, $user['hashed_password'])) {
            return $response->withStatus(500)->withJson(['error' => 'crypt error']);
        }

        $this->session->set('user_id', $user['id']);
        $bytes = random_bytes(20);
        $this->session->set('csrf_token', bin2hex($bytes));

        return $response->withJson(
            [
                'id' => $user['id'],
                'account_name' => $user['account_name'],
                'address' => $user['address'],
                'num_sell_items' => $user['num_sell_items'],
            ]
        );
    }

    public function settings(Request $request, Response $response, array $args)
    {
        $token = $this->session->get('csrf_token');

        try {
            $user = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(404)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        unset($user['hashed_password'], $user['last_bump'], $user['created_at'], $user['last_bump']);
        return $response->withStatus(200)->withJson(
            [
                'csrf_token' => $token,
                'user' => $user
            ]
        );
    }


    public function item(Request $request, Response $response, array $args)
    {
        $itemId = $args['id'];

        try {
            $user = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(404)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ?');
            $sth->execute([$itemId]);
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                return $response->withStatus(404)->withJson(['error' => 'item not found']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ?');
            $sth->execute([$item['seller_id']]);
            $seller = $sth->fetch(PDO::FETCH_ASSOC);
            if ($seller === false) {
                return $response->withStatus(404)->withJson(['error' => 'seller not found']);
            }

            unset($seller['hashed_password'], $seller['address'], $seller['created_at']);
            $item['seller'] = $seller;

            if (($user['id'] === $item['seller']['id'] || $user['id'] === $item['buyer_id']) && $item['buyer_id'] !== 0) {
                $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ?');
                $sth->execute([$item['buyer_id']]);
                $buyer = $sth->fetch(PDO::FETCH_ASSOC);
                if ($buyer === false) {
                    return $response->withStatus(404)->withJson(['error' => 'buyer not found']);
                }
                unset($buyer['hashed_password'], $buyer['address'], $buyer['created_at']);
                $item['buyer'] = $buyer;
            }
        } catch (\PDOException $e) {
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }
        unset($item['created_at'], $item['updated_at']);
        return $response->withStatus(200)->withJson($item);
    }


    public function sell(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(400)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(422)->withJson(['error' => 'csrf token error']);
        }

        // For test purpose, use 13 as default category
        $payload->category_id = $payload->category_id ?? 13;

        if (empty($payload->name) || empty($payload->description) || empty($payload->price) || $payload->price === 0 || empty($payload->category_id)) {
            return $response->withStatus(400)->withJson(['error' => 'all parameters are required']);
        }

        if ($payload->price < self::MIN_ITEM_PRICE || $payload->price > self::MAX_ITEM_PRICE) {
            return $response->withStatus(400)->withJson(['error' => '商品価格は100円以上、1,000,000円以下にしてください']);
        }

        $category = $this->getCategoryByID($payload->category_id);
        if ($category === false) {
            return $response->withStatus(400)->withJson(['error' => 'Incorrect category ID']);
        }

        try {
            $user = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(404)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        try {
            $this->dbh->beginTransaction();
            $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$user['id']]);
            $seller = $sth->fetch(PDO::FETCH_ASSOC);
            if ($seller === false) {
                $this->dbh->rollBack();
                $this->logger->warning('seller not found');
                return $response->withStatus(404)->withJson(['error' => 'user not found']);
            }

            $sth = $this->dbh->prepare('INSERT INTO `items` (`seller_id`, `status`, `name`, `price`, `description`, `category_id`) VALUES (?, ?, ?, ?, ?, ?)');
            $sth->execute([
                $seller['id'],
                self::ITEM_STATUS_ON_SALE,
                $payload->name,
                $payload->price,
                $payload->description,
                $payload->category_id
            ]);
            $itemId = $this->dbh->lastInsertId();

            $sth = $this->dbh->prepare('UPDATE `users` SET `num_sell_items`=?, `last_bump`=? WHERE `id`=?');
            $sth->execute([
                $seller['num_sell_items']+1,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $seller['id']
            ]);
        } catch (\PDOException $e) {
            $this->dbh->rollBack();
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        $this->dbh->commit();

        return $response->withStatus(200)->withJson(['id' => $itemId]);
    }

    public function edit(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(400)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(422)->withJson(['error' => 'csrf token error']);
        }

        if ($payload->price < self::MIN_ITEM_PRICE || $payload->price > self::MAX_ITEM_PRICE) {
            return $response->withStatus(400)->withJson(['error' => '商品価格は100円以上、1,000,000円以下にしてください']);
        }

        try {
            $user = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(404)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ?');
            $sth->execute([$payload->item_id]);
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->logger->warning('item not found', ['id' => $payload->item_id]);
                return $response->withStatus(404)->withJson(['error' => 'item not found']);
            }

            if ($item['seller_id'] !== $user['id']) {
                return $response->withStatus(403)->withJson(['error' => '自分の商品以外は編集できません']);
            }

            $this->dbh->beginTransaction();
            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$payload->item_id]);
            $item = $sth->fetch(PDO::FETCH_ASSOC);

            if ($item['status'] !== self::ITEM_STATUS_ON_SALE) {
                $this->dbh->rollBack();
                return $response->withStatus(403)->withJson(['error' => '販売中の商品以外編集できません']);
            }

            $sth = $this->dbh->prepare('UPDATE `items` SET `price` = ?, `updated_at` = ? WHERE `id` = ?');
            $sth->execute([$payload->price, (new \DateTime())->format(self::DATETIME_SQL_FORMAT), $payload->item_id]);
            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->dbh->rollBack();
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(200)->withJson([]);
    }

    public function buy(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(400)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(422)->withJson(['error' => 'csrf token error']);
        }

        try {
            $buyer = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(404)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        try {
            $this->dbh->beginTransaction();

            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$payload->item_id]);
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'item not found']);
            }

            if ($item['status'] !== self::ITEM_STATUS_ON_SALE) {
                $this->dbh->rollBack();
                return $response->withStatus(403)->withJson(['error' => 'item is not for sale']);
            }

            if ($item['seller_id'] === $buyer['id']) {
                $this->dbh->rollBack();
                return $response->withStatus(403)->withJson(['error' => '自分の商品は買えません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$item['seller_id']]);
            $seller = $sth->fetch(PDO::FETCH_ASSOC);
            if ($seller === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'seller not found']);
            }

            $category = $this->getCategoryByID($item['category_id']);
            if ($category === false) {
                return $response->withStatus(500)->withJson(['error' => 'category id error']);
            }

            $sth = $this->dbh->prepare('INSERT INTO `transaction_evidences` '.
                '(`seller_id`, `buyer_id`, `status`, '.
                '`item_id`, `item_name`, `item_price`, `item_description`, '.
                '`item_category_id`, `item_root_category_id`) '.
                'VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)');
            $sth->execute([
                $item['seller_id'],
                $buyer['id'],
                self::TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING,
                $item['id'],
                $item['name'],
                $item['price'],
                $item['description'],
                $category['id'],
                $category['parent_id'],
            ]);
            $transactionEvidenceId = $this->dbh->lastInsertId();

            $sth = $this->dbh->prepare('UPDATE `items` SET `buyer_id` = ?, `status` = ?, `updated_at` = ? WHERE `id` = ?');
            $sth->execute([
                $buyer['id'],
                self::ITEM_STATUS_TRADING,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $item['id'],
            ]);

            $client = new Client();
            $res = $client->post(
                'http://localhost:7000/create',
                [
                    'headers' => ['Authorization' => self::ISUCARI_API_TOKEN],
                    'json' => [
                        'to_address' => $buyer['address'],
                        'to_name' => $buyer['account_name'],
                        'from_address' => $seller['address'],
                        'from_name' => $seller['account_name'],
                    ]
                ]
            );
            if ($res->getStatusCode() != 200) {
                $this->dbh->rollBack();
                $this->logger->error($res->getReasonPhrase());
                return $response->withStatus(500)->withJson(['error' => 'failed to request to shipment service']);
            }
            $shippingResponse = json_decode($res->getBody());

            $pres = $client->post(
                'http://localhost:5555/token',
                ['json' => [
                    'shop_id' => self::PAYMENT_SERVICE_ISUCARI_SHOP_ID,
                    'api_key' => self::PAYMENT_SERVICE_ISUCARI_API_KEY,
                    'token' =>  $payload->token,
                    'price' => $item['price'],
                ]]
            );

            if ($pres->getStatusCode() != 200) {
                $this->dbh->rollBack();
                $this->logger->error($res->getReasonPhrase());
                return $response->withStatus(500)->withJson(['error' => 'payment service is failed']);
            }

            $paymentResponse = json_decode($pres->getBody());
            if (json_last_error() !== JSON_ERROR_NONE) {
                $this->dbh->rollBack();
                $this->logger->error(json_last_error_msg());
                return $response->withStatus(500)->withJson(['error' => 'payment service is failed']);
            }

            if ($paymentResponse->status === 'invalid') {
                $this->dbh->rollBack();
                return $response->withStatus(400)->withJson(['error' => 'カード情報に誤りがあります']);
            }

            if ($paymentResponse->status === 'fail') {
                $this->dbh->rollBack();
                return $response->withStatus(400)->withJson(['error' => 'カードの残高が足りません']);
            }

            if ($paymentResponse->status !== 'ok') {
                $this->dbh->rollBack();
                return $response->withStatus(400)->withJson(['error' => '想定外のエラー']);
            }

            $sth = $this->dbh->prepare('INSERT INTO `shippings` '.
                '(`transaction_evidence_id`, `status`, `item_name`, `item_id`, `reserve_id`, `reserve_time`, '.
                '`to_address`, `to_name`, `from_address`, `from_name`, `img_name`) '.
                ' VALUES (?,?,?,?,?,?,?,?,?,?,?)');
            $sth->execute([
                $transactionEvidenceId,
                self::SHIPPING_STATUS_INITIAL,
                $item['name'],
                $item['id'],
                $shippingResponse->reserve_id,
                $shippingResponse->reserve_time,
                $buyer['address'],
                $buyer['account_name'],
                $seller['address'],
                $seller['account_name'],
                ""
            ]);

            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->dbh->rollBack();
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(200)->withJson(['transaction_evidence_id' => $transactionEvidenceId]);
    }

    public function ship(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(400)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(422)->withJson(['error' => 'csrf token error']);
        }

        try {
            $seller = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(404)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `item_id` = ?');
            $sth->execute([$payload->item_id]);
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                return $response->withStatus(404)->withJson(['error' => 'transaction_evidences not found']);
            }

            if ($transactionEvidence['seller_id'] !== $seller['id']) {
                return $response->withStatus(403)->withJson(['error' => '権限がありません']);
            }

            $this->dbh->beginTransaction();
            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$payload->item_id]);
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'item not found']);
            }

            if ($item['status'] !== self::ITEM_STATUS_TRADING) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => '商品が取引中ではありません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$transactionEvidence['id']]);
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'transaction_evidences not found']);
            }

            if ($transactionEvidence['status'] !== self::TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING) {
                $this->dbh->rollBack();
                return $response->withStatus(403)->withJson(['error' => '準備ができていません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE');
            $sth->execute([$transactionEvidence['id']]);
            $shipping = $sth->fetch(PDO::FETCH_ASSOC);
            if ($shipping === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'shippings not found']);
            }

            $bytes = random_bytes(16);
            $imgName = bin2hex($bytes);
            $path = $this->settings['app']['upload_path'] . $imgName . '.png';
            $resource = fopen($path, 'w');

            $client = new \GuzzleHttp\Client();
            $r = $client->post(
                'http://localhost:7000/request',
                [
                    'headers' => ['Authorization' => self::ISUCARI_API_TOKEN],
                    'json' => ['reserve_id' => $shipping['reserve_id']],
                    'sink' => $resource,
                ]
            );
            fclose($resource);
            if ($r->getStatusCode() !== 200) {
                $this->logger->error($r->getReasonPhrase());
                $this->dbh->rollBack();
                return $response->withStatus(500)->withJson(['error' => 'failed to request to shipment service']);
            }


            $sth = $this->dbh->prepare('UPDATE `shippings` SET `status` = ?, `img_name` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?');
            $sth->execute([
                self::SHIPPING_STATUS_WAIT_PICKUP,
                $imgName,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $transactionEvidence['id']
            ]);

            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(200)->withJson([
            'url' => sprintf("http://%s/upload/%s.png", $request->getServerParam('HTTP_HOST'), $imgName),
        ]);
    }

    public function ship_done(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(400)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(422)->withJson(['error' => 'csrf token error']);
        }

        try {
            $seller = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(404)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `item_id` = ?');
            $sth->execute([$payload->item_id]);
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                return $response->withStatus(404)->withJson(['error' => 'transaction_evidence not found']);
            }

            if ($transactionEvidence['seller_id'] !== $seller['id']) {
                return $response->withStatus(403)->withJson(['error' => '権限がありません']);
            }

            $this->dbh->beginTransaction();

            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$payload->item_id]);
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'item not found']);
            }

            if ($item['status'] != self::ITEM_STATUS_TRADING) {
                $this->dbh->rollBack();
                return $response->withStatus(403)->withJson(['error' => '商品が取引中ではありません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$transactionEvidence['id']]);
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'transaction_evidences not found']);
            }

            if ($transactionEvidence['status'] !== self::TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING) {
                $this->dbh->rollBack();
                return $response->withStatus(403)->withJson(['error' => '準備ができていません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE');
            $sth->execute([$transactionEvidence['id']]);
            $shipping = $sth->fetch(PDO::FETCH_ASSOC);
            if ($shipping === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'shippings not found']);
            }

            $client = new Client();
            $r = $client->get('http://localhost:7000/status', [
                'headers' => ['Authorization' => self::ISUCARI_API_TOKEN],
                'json' => ['reserve_id' => $shipping['reserve_id']],
            ]);
            if ($r->getStatusCode() !== 200) {
                $this->logger->error($r->getReasonPhrase());
                $this->dbh->rollBack();
                return $response->withStatus(500)->withJson(['error' => 'failed to request to shipment service']);
            }
            $shippingResponse = json_decode($r->getBody());
            if (! ($shippingResponse->status === self::SHIPPING_STATUS_DONE || $shippingResponse->status === self::SHIPPING_STATUS_SHIPPING)) {
                $this->dbh->rollBack();
                return $response->withStatus(500)->withJson(['error' => 'shipment service側で配送中か配送完了になっていません']);
            }

            $sth = $this->dbh->prepare('UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?');
            $sth->execute([
                $shippingResponse->status,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $transactionEvidence['id'],
            ]);

            $sth = $this->dbh->prepare('UPDATE `transaction_evidences` SET `status` = ?, `updated_at` = ? WHERE `id` = ?');
            $sth->execute([
                self::TRANSACTION_EVIDENCE_STATUS_WAIT_DONE,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $transactionEvidence['id'],
            ]);

            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(200)->withJson([]);
    }

    public function complete(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(400)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(422)->withJson(['error' => 'csrf token error']);
        }

        try {
            $buyer = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(404)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        try {
            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `item_id` = ?');
            $sth->execute([$payload->item_id]);
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                return $response->withStatus(404)->withJson(['error' => 'transaction_evidence not found']);
            }

            if ($transactionEvidence['buyer_id'] !== $buyer['id']) {
                return $response->withStatus(403)->withJson(['error' => '権限がありません']);
            }

            $this->dbh->beginTransaction();

            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$payload->item_id]);
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'item not found']);
            }

            if ($item['status'] != self::ITEM_STATUS_TRADING) {
                $this->dbh->rollBack();
                return $response->withStatus(403)->withJson(['error' => '商品が取引中ではありません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$transactionEvidence['id']]);
            $transactionEvidence = $sth->fetch(PDO::FETCH_ASSOC);
            if ($transactionEvidence === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'transaction_evidences not found']);
            }

            if ($transactionEvidence['status'] !== self::TRANSACTION_EVIDENCE_STATUS_WAIT_DONE) {
                $this->dbh->rollBack();
                return $response->withStatus(403)->withJson(['error' => '準備ができていません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE');
            $sth->execute([$transactionEvidence['id']]);
            $shipping = $sth->fetch(PDO::FETCH_ASSOC);
            if ($shipping === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'shippings not found']);
            }

            $client = new Client();
            $r = $client->post('http://localhost:7000/status', [
                'headers' => ['Authorization' => self::ISUCARI_API_TOKEN],
                'json' => ['reserve_id' => $shipping['reserve_id']],
            ]);
            if ($r->getStatusCode() !== 200) {
                $this->logger->error($r->getReasonPhrase());
                $this->dbh->rollBack();
                return $response->withStatus(500)->withJson(['error' => 'failed to request to shipment service']);
            }
            $shippingResponse = json_decode($r->getBody());
            if ($shippingResponse->status !== self::SHIPPING_STATUS_DONE) {
                $this->dbh->rollBack();
                return $response->withStatus(500)->withJson(['error' => 'shipment service側で配送完了になっていません']);
            }

            $sth = $this->dbh->prepare('UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?');
            $sth->execute([
                self::SHIPPING_STATUS_DONE,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $transactionEvidence['id'],
            ]);

            $sth = $this->dbh->prepare('UPDATE `transaction_evidences` SET `status` = ?, `updated_at` = ? WHERE `id` = ?');
            $sth->execute([
                self::TRANSACTION_EVIDENCE_STATUS_DONE,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $transactionEvidence['id'],
            ]);

            $sth = $this->dbh->prepare('UPDATE `items` SET `status` = ?, `updated_at` = ? WHERE `id` = ?');
            $sth->execute([
                self::ITEM_STATUS_SOLD_OUT,
                (new \DateTime())->format(self::DATETIME_SQL_FORMAT),
                $item['id'],
            ]);

            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(200)->withJson([]);
    }

    public function bump(Request $request, Response $response, array $args)
    {
        try {
            $payload = $this->jsonPayload($request);
        } catch (\InvalidArgumentException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(400)->withJson(['error' => 'json decode error']);
        }

        if ($payload->csrf_token !== $this->session->get('csrf_token')) {
            return $response->withStatus(422)->withJson(['error' => 'csrf token error']);
        }

        try {
            $user = $this->getCurrentUser();
        } catch (\DomainException $e) {
            $this->logger->warning('user not found');
            return $response->withStatus(404)->withJson(['error' => 'user not found']);
        } catch (\Exception $e) {
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        try {
            $this->dbh->beginTransaction();

            $sth = $this->dbh->prepare('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$payload->item_id]);
            $item = $sth->fetch(PDO::FETCH_ASSOC);
            if ($item === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'item not found']);
            }

            if ($item['seller_id'] !== $user['id']) {
                $this->dbh->rollBack();
                return $response->withStatus(403)->withJson(['error' => '自分の商品以外は編集できません']);
            }

            $sth = $this->dbh->prepare('SELECT * FROM `users` WHERE `id` = ? FOR UPDATE');
            $sth->execute([$user['id']]);
            $seller = $sth->fetch(PDO::FETCH_ASSOC);
            if ($seller === false) {
                $this->dbh->rollBack();
                return $response->withStatus(404)->withJson(['error' => 'user not found']);
            }

            // last_bump + 3s > now
            $now = new \DateTime();
            if ((new \DateTime($seller['last_bump']))->getTimestamp() + self::BUMP_CHARGE_SECONDS > $now->getTimestamp()) {
                $this->dbh->rollBack();
                return $response->withStatus(400)->withJson(['error' => 'Bump not allowed']);
            }

            $sth = $this->dbh->prepare('UPDATE `items` SET `created_at`=? WHERE id=?');
            $sth->execute([
                $now->format(self::DATETIME_SQL_FORMAT),
                $item['id']
            ]);

            $sth = $this->dbh->prepare('UPDATE `users` SET `last_bump`=? WHERE id=?');
            $sth->execute([
                $now->format(self::DATETIME_SQL_FORMAT),
                $user['id']
            ]);

            $this->dbh->commit();
        } catch (\PDOException $e) {
            $this->logger->error($e->getMessage());
            return $response->withStatus(500)->withJson(['error' => 'db error']);
        }

        return $response->withStatus(200)->withJson([]);
    }
}
