package common

var TemplateLetter = `
		<div style="background-color: #0a0a0c; padding: 50px 20px; font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;">

			<div style="max-width: 480px; margin: 0 auto; background-color: #131318; padding: 40px 30px; border-radius: 16px; border: 1px solid #2a2a35; border-top: 4px solid #8b5cf6; text-align: center;">

				<div style="font-size: 32px; font-weight: bold; margin-bottom: 30px; letter-spacing: 1px;">
					<span style="color: #ffffff;">Ne</span><span style="color: #8b5cf6;">X</span><span style="color: #ffffff;">uS</span>
				</div>

				<h2 style="color: #ffffff; margin-bottom: 15px; font-size: 22px; font-weight: 500;">
					Восстановление пароля
				</h2>

				<p style="color: #a1a1aa; font-size: 15px; line-height: 1.6; margin-bottom: 30px;">
					Вы запросили сброс пароля.<br>Введите этот код на сайте для подтверждения:
				</p>

				<div style="background-color: #1a1528; border: 1px solid #5a32a3; border-radius: 12px; padding: 20px 20px 20px 32px; margin: 0 auto 30px auto; display: inline-block;">
					<div style="font-size: 40px; font-weight: bold; letter-spacing: 12px; color: #a78bfa;">
						%s
					</div>
				</div>

				<p style="color: #52525b; font-size: 13px; line-height: 1.5; margin-top: 10px;">
					Код действует 15 минут.<br>
					Если вы не запрашивали сброс пароля, просто проигнорируйте это письмо. Никому не сообщайте данный код.
				</p>

			</div>

		</div>
	`
