<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet version="1.0"
	xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
	xmlns:atom="http://www.w3.org/2005/Atom"
	exclude-result-prefixes="atom">
	<xsl:output method="html" encoding="UTF-8" indent="yes" doctype-system="about:legacy-compat"/>

	<xsl:template match="/atom:feed">
		<html lang="en">
			<head>
				<title><xsl:value-of select="atom:title"/> - Gentoo Packages</title>
				<meta charset="utf-8"/>
				<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
				<meta name="theme-color" content="#54487a"/>
				<meta name="description" content="Gentoo Packages Database"/>
				<link rel="icon" href="https://packages.gentoo.org/favicon.ico" type="image/x-icon"/>
				<link rel="stylesheet" href="/assets/stylesheets.css"/>
			</head>
			<body>
				<header>
					<div class="site-title">
						<div class="container">
							<div class="row justify-content-between">
								<div class="logo">
									<a href="/" title="Back to the homepage" class="site-logo">
										<img src="https://assets.gentoo.org/tyrian/site-logo.png" alt="Gentoo" srcset="https://assets.gentoo.org/tyrian/site-logo.svg"/>
									</a>
									<span class="site-label">Packages</span>
								</div>
							</div>
						</div>
					</div>
					<nav class="tyrian-navbar navbar navbar-dark navbar-expand bg-primary" role="navigation">
						<div class="container">
							<ul class="navbar-nav mr-auto">
								<li class="nav-item"><a class="nav-link" href="/">Home</a></li>
								<li class="nav-item"><a class="nav-link" href="/categories">Packages</a></li>
								<li class="nav-item"><a class="nav-link" href="/maintainers">Maintainers</a></li>
								<li class="nav-item"><a class="nav-link" href="/useflags">USE flags</a></li>
								<li class="nav-item"><a class="nav-link" href="/arches">Architectures</a></li>
								<li class="nav-item"><a class="nav-link" href="/about">About</a></li>
							</ul>
						</div>
					</nav>
				</header>
				<div class="container mb-5">
					<div class="row">
						<div class="col-12">
							<h1 class="stick-top mt-4 mb-2">
								<span class="fa fa-fw fa-rss-square kk-feed-icon"></span>
								<xsl:value-of select="atom:title"/>
							</h1>
							<xsl:if test="atom:subtitle and atom:subtitle != atom:title">
								<p class="lead"><xsl:value-of select="atom:subtitle"/></p>
							</xsl:if>
							<div class="alert alert-info">
								<strong><span class="fa fa-fw fa-rss"></span> This is a web feed.</strong>
								<br/>
								It is meant to be read by a feed reader, which will tell you whenever
								something new shows up here. To subscribe, paste the address of this
								page into your reader of choice.
								<br/>
								<small>
									<a class="alert-link" href="https://aboutfeeds.com/">New to feeds? Here is a short introduction.</a>
								</small>
							</div>
						</div>
						<div class="col-12">
							<xsl:choose>
								<xsl:when test="atom:entry">
									<ul class="list-group">
										<xsl:apply-templates select="atom:entry"/>
									</ul>
								</xsl:when>
								<xsl:otherwise>
									<p class="text-muted">This feed is currently empty.</p>
								</xsl:otherwise>
							</xsl:choose>
						</div>
					</div>
				</div>
				<footer style="background-color: #fafafa; box-shadow:none!important;">
					<div class="container pt-4" style="border-top: 1px solid #dddddd;">
						<div class="row">
							<div class="col-12">
								<strong>© 2001-2026 Gentoo Authors</strong>
								<br/>
								<small>
									Gentoo is a trademark of the Gentoo Foundation, Inc. and of Förderverein Gentoo e.V.
									The contents of this document, unless otherwise expressly stated, are licensed under the
									<a href="https://creativecommons.org/licenses/by-sa/4.0/" rel="license">CC-BY-SA-4.0</a> license.
									The <a href="https://www.gentoo.org/inside-gentoo/foundation/name-logo-guidelines.html">Gentoo Name and Logo Usage Guidelines</a> apply.
								</small>
							</div>
						</div>
					</div>
				</footer>
			</body>
		</html>
	</xsl:template>

	<xsl:template match="atom:entry">
		<li class="list-group-item">
			<div class="row">
				<div class="col-xs-12 col-md-9">
					<h4 class="mb-1">
						<a href="{atom:link[@rel='alternate']/@href | atom:link[not(@rel)]/@href}">
							<xsl:call-template name="unescape">
								<xsl:with-param name="text" select="atom:title"/>
							</xsl:call-template>
						</a>
					</h4>
					<xsl:if test="atom:summary != ''">
						<p class="mb-1">
							<xsl:call-template name="unescape">
								<xsl:with-param name="text" select="atom:summary"/>
							</xsl:call-template>
						</p>
					</xsl:if>
				</div>
				<div class="col-xs-12 col-md-3 text-md-right">
					<small class="text-muted">
						<xsl:value-of select="substring(atom:updated, 1, 10)"/>
						<xsl:if test="atom:author/atom:name != ''">
							<br/>
							<xsl:value-of select="atom:author/atom:name"/>
						</xsl:if>
					</small>
				</div>
			</div>
		</li>
	</xsl:template>

	<!-- Titles and summaries are declared as type="html", so their text is HTML
	     escaped on top of the XML escaping. Undo one level and emit the result as
	     text; disable-output-escaping is no help, browsers ignore it because they
	     build the result tree directly instead of serializing it. -->
	<xsl:template name="unescape">
		<xsl:param name="text"/>
		<xsl:variable name="lt">
			<xsl:call-template name="replace">
				<xsl:with-param name="text" select="string($text)"/>
				<xsl:with-param name="from" select="'&amp;lt;'"/>
				<xsl:with-param name="to" select="'&lt;'"/>
			</xsl:call-template>
		</xsl:variable>
		<xsl:variable name="gt">
			<xsl:call-template name="replace">
				<xsl:with-param name="text" select="string($lt)"/>
				<xsl:with-param name="from" select="'&amp;gt;'"/>
				<xsl:with-param name="to" select="'&gt;'"/>
			</xsl:call-template>
		</xsl:variable>
		<xsl:variable name="quot">
			<xsl:call-template name="replace">
				<xsl:with-param name="text" select="string($gt)"/>
				<xsl:with-param name="from" select="'&amp;#34;'"/>
				<xsl:with-param name="to" select="'&quot;'"/>
			</xsl:call-template>
		</xsl:variable>
		<xsl:variable name="apos">
			<xsl:call-template name="replace">
				<xsl:with-param name="text" select="string($quot)"/>
				<xsl:with-param name="from" select="'&amp;#39;'"/>
				<xsl:with-param name="to" select="&quot;'&quot;"/>
			</xsl:call-template>
		</xsl:variable>
		<!-- last, so it cannot turn an escaped &amp;lt; into a < of its own -->
		<xsl:call-template name="replace">
			<xsl:with-param name="text" select="string($apos)"/>
			<xsl:with-param name="from" select="'&amp;amp;'"/>
			<xsl:with-param name="to" select="'&amp;'"/>
		</xsl:call-template>
	</xsl:template>

	<xsl:template name="replace">
		<xsl:param name="text"/>
		<xsl:param name="from"/>
		<xsl:param name="to"/>
		<xsl:choose>
			<xsl:when test="contains($text, $from)">
				<xsl:value-of select="substring-before($text, $from)"/>
				<xsl:value-of select="$to"/>
				<xsl:call-template name="replace">
					<xsl:with-param name="text" select="substring-after($text, $from)"/>
					<xsl:with-param name="from" select="$from"/>
					<xsl:with-param name="to" select="$to"/>
				</xsl:call-template>
			</xsl:when>
			<xsl:otherwise>
				<xsl:value-of select="$text"/>
			</xsl:otherwise>
		</xsl:choose>
	</xsl:template>
</xsl:stylesheet>
